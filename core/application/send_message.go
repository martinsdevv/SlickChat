package application

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/martinsdevv/slickchat/core/events"
	kafkainfra "github.com/martinsdevv/slickchat/infrastructure/kafka"
)

func SendMessage(producer *kafkainfra.Producer, userID, roomID, content string) (string, error) {
	ctx := context.Background()
	messageID := uuid.New().String()

	payload := events.MessageSent{
		MessageID:        messageID,
		RoomID:           roomID,
		SenderID:         userID,
		Content:          content,
		MessageType:      "TEXT",
		IsZeroLogging:    false,
		TTL:              0,
		DestroyAfterRead: false,
		ExpiresAt:        nil,
		Timestamp:        time.Now().UTC(),
	}

	event, err := events.NewEvent(
		events.EventTypeMessageSent,
		roomID,
		payload,
	)
	if err != nil {
		return "", err
	}

	err = producer.Publish(ctx, event)

	if err != nil {
		return messageID, err
	}

	return messageID, nil
}
