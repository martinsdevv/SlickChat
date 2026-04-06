package application

import (
	"context"
	"time"

	"github.com/martinsdevv/slickchat/core/events"
	kafkainfra "github.com/martinsdevv/slickchat/infrastructure/kafka"
)

func ReadMessage(producer *kafkainfra.Producer, userID, roomID, messageID string) error {
	payload := events.MessageRead{
		MessageID: messageID,
		RoomID:    roomID,
		UserID:    userID,
		ReadAt:    time.Now().UTC(),
	}

	event, err := events.NewEvent(
		events.EventTypeMessageRead,
		roomID,
		payload,
	)

	if err != nil {
		return err
	}

	return producer.Publish(context.Background(), event)
}
