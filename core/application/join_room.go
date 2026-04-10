package application

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/martinsdevv/slickchat/core/contracts"
	"github.com/martinsdevv/slickchat/core/domain"
	"github.com/martinsdevv/slickchat/core/events"
	kafkainfra "github.com/martinsdevv/slickchat/infrastructure/kafka"
)

var ErrRoomNotFound = errors.New("room not found")

func JoinRoom(
	ctx context.Context,
	producer *kafkainfra.Producer,
	rooms contracts.RoomRepository,
	memberships contracts.RoomMembershipRepository,
	roomID uuid.UUID,
	userID uuid.UUID,
	role domain.Role,
) error {
	if _, err := rooms.GetByID(ctx, roomID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrRoomNotFound
		}
		return err
	}

	if err := memberships.Add(ctx, roomID, userID, role); err != nil {
		return err
	}

	now := time.Now().UTC()
	payload := events.UserJoinedRoom{
		RoomID:   roomID.String(),
		UserID:   userID.String(),
		JoinedAt: now,
		Role:     string(role),
	}

	event, err := events.NewEvent(events.EventTypeUserJoinedRoom, roomID.String(), payload)
	if err != nil {
		return err
	}

	return producer.Publish(ctx, event)
}
