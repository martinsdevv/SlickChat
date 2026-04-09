package application

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/martinsdevv/slickchat/core/events"
	kafkainfra "github.com/martinsdevv/slickchat/infrastructure/kafka"
)

func PublishMessageExpired(
	producer *kafkainfra.Producer,
	messageID uuid.UUID,
	roomID uuid.UUID,
	expiredAt time.Time,
) error {
	payload := events.MessageExpired{
		MessageID: messageID.String(),
		RoomID:    roomID.String(),
		ExpiredAt: expiredAt.UTC(),
	}

	event, err := events.NewEvent(
		events.EventTypeMessageExpired,
		roomID.String(),
		payload,
	)
	if err != nil {
		return err
	}

	return producer.Publish(context.Background(), event)
}
