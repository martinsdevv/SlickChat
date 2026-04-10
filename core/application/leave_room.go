package application

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/martinsdevv/slickchat/core/contracts"
	"github.com/martinsdevv/slickchat/core/events"
	kafkainfra "github.com/martinsdevv/slickchat/infrastructure/kafka"
)

var ErrNotRoomMember = errors.New("not a room member")

func LeaveRoom(
	ctx context.Context,
	producer *kafkainfra.Producer,
	memberships contracts.RoomMembershipRepository,
	roomID uuid.UUID,
	userID uuid.UUID,
) error {
	if _, err := memberships.Get(ctx, roomID, userID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotRoomMember
		}
		return err
	}

	if err := memberships.Remove(ctx, roomID, userID); err != nil {
		return err
	}

	now := time.Now().UTC()
	payload := events.UserLeftRoom{
		RoomID: roomID.String(),
		UserID: userID.String(),
		LeftAt: now,
	}

	event, err := events.NewEvent(events.EventTypeUserLeftRoom, roomID.String(), payload)
	if err != nil {
		return err
	}

	return producer.Publish(ctx, event)
}
