package fanout

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

type MessageSent struct {
	MessageID string    `json:"message_id"`
	RoomID    string    `json:"room_id"`
	SenderID  string    `json:"sender_id"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

func StartConsumer(broker string, handler func(MessageSent)) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{broker},
		Topic:   "message-events",
		GroupID: "fanout-group",
	})

	for {
		msg, err := reader.ReadMessage(context.Background())
		if err != nil {
			log.Println(err)
			continue
		}

		var event MessageSent
		json.Unmarshal(msg.Value, &event)

		handler(event)
	}
}
