package events

import "time"

const (
	EventTypeMessageSent      = "message.sent.v1"
	EventTypeMessageDelivered = "message.delivered.v1"
	EventTypeMessageRead      = "message.read.v1"
	EventTypeMessageDeleted   = "message.deleted.v1"
)

type MessageSent struct {
	MessageID        string     `json:"message_id"`
	RoomID           string     `json:"room_id"`
	SenderID         string     `json:"sender_id"`
	MessageType      string     `json:"message_type"`
	Content          string     `json:"content"`
	IsZeroLogging    bool       `json:"is_zero_logging"`
	TTL              int        `json:"ttl"`
	DestroyAfterRead bool       `json:"destroy_after_read"`
	ExpiresAt        *time.Time `json:"expires_at,omitempty"`
	SentAt           time.Time  `json:"sent_at"`
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

type MessageDeleted struct {
	MessageID string    `json:"message_id"`
	RoomID    string    `json:"room_id"`
	DeletedAt time.Time `json:"deleted_at"`
}
