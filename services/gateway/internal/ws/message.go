package ws

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/martinsdevv/slickchat/core/events"
	kafkainfra "github.com/martinsdevv/slickchat/infrastructure/kafka"
)

func handleSendMessage(producer *kafkainfra.Producer, payload SendMessagePayload, userId string) {
	payloadRaw := events.MessageSent{
		MessageID: uuid.New().String(),
		RoomID:    payload.RoomID,
		SenderID:  userId,
		Content:   payload.Content,
	}

	payloadBytes, _ := json.Marshal(payloadRaw)

	event := events.Event{
		ID:        uuid.New().String(),
		Type:      events.EventTypeMessageSent,
		Timestamp: time.Now(),
		Payload:   payloadBytes,
	}

	producer.Publish(context.Background(), event)
}

func sendAck(client *Client) {
	client.Write(map[string]interface{}{
		"type": "message_ack",
		"payload": map[string]string{
			"status": "received",
		},
	})
}
