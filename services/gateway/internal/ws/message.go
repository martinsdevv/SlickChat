package ws

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	redisclient "github.com/martinsdevv/slickchat/services/gateway/internal/redis"
	"github.com/redis/go-redis/v9"
)

type MessageSent struct {
	MessageID string    `json:"message_id"`
	RoomID    string    `json:"room_id"`
	SenderID  string    `json:"sender_id"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

func handleSendMessage(rdb *redis.Client, payload SendMessagePayload) {
	event := MessageSent{
		MessageID: uuid.New().String(),
		RoomID:    payload.RoomID,
		SenderID:  "1",
		Content:   payload.Content,
		Timestamp: time.Now(),
	}

	data, _ := json.Marshal(event)

	rdb.Publish(redisclient.Ctx, "room:"+payload.RoomID, data)
}

func sendAck(client *Client) {
	client.Write(map[string]interface{}{
		"type": "message_ack",
		"payload": map[string]string{
			"status": "received",
		},
	})
}
