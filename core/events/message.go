package events

const (
	EventTypeMessageSent = "message.sent.v1"
)

type MessageSent struct {
	MessageID string `json:"message_id"`
	RoomID    string `json:"room_id"`
	SenderID  string `json:"sender_id"`
	Content   string `json:"content"`
}
