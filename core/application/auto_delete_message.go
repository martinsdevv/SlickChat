package application

import (
	"context"
	"time"

	"github.com/martinsdevv/slickchat/core/events"
	kafkainfra "github.com/martinsdevv/slickchat/infrastructure/kafka"
)

// AutoDeleteMessage publica MessageDeleted sem checagem de permissão.
// Usado para comportamentos sistêmicos (ex: destroy-after-read).
func AutoDeleteMessage(producer *kafkainfra.Producer, roomID, messageID string) error {
	payload := events.MessageDeleted{
		MessageID: messageID,
		RoomID:    roomID,
		DeletedAt: time.Now().UTC(),
	}

	event, err := events.NewEvent(
		events.EventTypeMessageDeleted,
		roomID,
		payload,
	)
	if err != nil {
		return err
	}

	return producer.Publish(context.Background(), event)
}

