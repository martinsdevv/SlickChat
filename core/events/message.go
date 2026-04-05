package events

import "time"

const (
	EventTypeMessageSent      = "message.sent.v1"
	EventTypeMessageDelivered = "message.delivered.v1"
	EventTypeMessageRead      = "message.read.v1"
)

type MessageSent struct {
	MessageID string `json:"message_id"`
	RoomID    string `json:"room_id"`
	SenderID  string `json:"sender_id"`
	Content   string `json:"content"`
}

type MessageDelivered struct {
	MessageID   string    `json:"message_id"`
	RoomID      string    `json:"room_id"`
	UserID      string    `json:"user_id"`
	DeliveredAt time.Time `json:"delivered_at"`
}

type MessageRead struct {
	MessageID string    `json:"message_id"`
	RoomID    string    `json:"room_id"`
	UserID    string    `json:"user_id"`
	ReadAt    time.Time `json:"read_at"`
}
