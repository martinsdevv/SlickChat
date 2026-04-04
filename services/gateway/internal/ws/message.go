package ws

import (
	"context"
	"time"

	"github.com/google/uuid"
	kafkainfra "github.com/martinsdevv/slickchat/infrastructure/kafka"
)

type MessageSent struct {
	MessageID string    `json:"message_id"`
	RoomID    string    `json:"room_id"`
	SenderID  string    `json:"sender_id"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

func handleSendMessage(producer *kafkainfra.Producer, payload SendMessagePayload) {
	event := MessageSent{
		MessageID: uuid.New().String(),
		RoomID:    payload.RoomID,
		SenderID:  "1",
		Content:   payload.Content,
		Timestamp: time.Now(),
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
