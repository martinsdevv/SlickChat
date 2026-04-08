package application

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/martinsdevv/slickchat/core/domain"
	"github.com/martinsdevv/slickchat/core/events"
	kafkainfra "github.com/martinsdevv/slickchat/infrastructure/kafka"
)

func DeleteMessage(
	producer *kafkainfra.Producer,
	room *domain.Room,
	membership *domain.RoomMembership,
	userID uuid.UUID,
	messageID uuid.UUID,
	messageSenderID uuid.UUID,
) error {

	now := time.Now().UTC()

	if err := room.CanDeleteMessage(
		userID,
		membership,
		messageSenderID,
		now,
	); err != nil {
		return err
	}

	payload := events.MessageDeleted{
		MessageID: messageID.String(),
		RoomID:    room.ID.String(),
		DeletedAt: now,
	}

	event, err := events.NewEvent(
		events.EventTypeMessageDeleted,
		room.ID.String(),
		payload,
	)
	if err != nil {
		return err
	}

	return producer.Publish(context.Background(), event)
}
