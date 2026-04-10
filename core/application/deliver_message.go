package application

import (
	"context"
	"time"

	"github.com/martinsdevv/slickchat/core/events"
	kafkainfra "github.com/martinsdevv/slickchat/infrastructure/kafka"
)

func DeliverMessage(
	producer *kafkainfra.Producer,
	userID, roomID, messageID, senderID string,
) error {

	now := time.Now().UTC()

	payload := events.MessageDelivered{
		MessageID:   messageID,
		RoomID:      roomID,
		UserID:      userID,
		SenderID:    senderID,
		DeliveredAt: now,
	}

	event, err := events.NewEvent(
		events.EventTypeMessageDelivered,
		roomID,
		payload,
	)
	if err != nil {
		return err
	}

	return producer.Publish(context.Background(), event)
}
