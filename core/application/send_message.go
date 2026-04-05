package application

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/martinsdevv/slickchat/core/events"
	kafkainfra "github.com/martinsdevv/slickchat/infrastructure/kafka"
)

func SendMessage(producer *kafkainfra.Producer, userID, roomID, content string) error {
	payload := events.MessageSent{
		MessageID: uuid.New().String(),
		RoomID:    roomID,
		SenderID:  userID,
		Content:   content,
	}

	payloadBytes, _ := json.Marshal(payload)

	event := events.Event{
		ID:        uuid.New().String(),
		Type:      events.EventTypeMessageSent,
		Timestamp: time.Now(),
		Payload:   payloadBytes,
	}

	return producer.Publish(context.Background(), event)
}
