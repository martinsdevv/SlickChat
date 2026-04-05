package application

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/martinsdevv/slickchat/core/events"
	kafkainfra "github.com/martinsdevv/slickchat/infrastructure/kafka"
)

func DeliverMessage(producer *kafkainfra.Producer, userID, roomID, messageID string) error {
	payload := events.MessageDelivered{
		MessageID:   messageID,
		RoomID:      roomID,
		UserID:      userID,
		DeliveredAt: time.Now(),
	}

	payloadBytes, _ := json.Marshal(payload)

	event := events.Event{
		ID:        uuid.New().String(),
		Type:      events.EventTypeMessageDelivered,
		Timestamp: time.Now(),
		Payload:   payloadBytes,
	}

	return producer.Publish(context.Background(), event)
}
