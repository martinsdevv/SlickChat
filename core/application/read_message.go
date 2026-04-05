package application

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/martinsdevv/slickchat/core/events"
	kafkainfra "github.com/martinsdevv/slickchat/infrastructure/kafka"
)

func ReadMessage(producer *kafkainfra.Producer, userID, roomID, messageID string) error {
	payload := events.MessageRead{
		MessageID: messageID,
		RoomID:    roomID,
		UserID:    userID,
		ReadAt:    time.Now(),
	}

	payloadBytes, _ := json.Marshal(payload)

	event := events.Event{
		ID:        uuid.New().String(),
		Type:      events.EventTypeMessageRead,
		Timestamp: time.Now(),
		Payload:   payloadBytes,
	}

	return producer.Publish(context.Background(), event)
}
