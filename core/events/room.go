package events

import "time"

const (
	EventTypeUserJoinedRoom = "room.user_joined.v1"
	EventTypeUserLeftRoom   = "room.user_left.v1"
)

type UserJoinedRoom struct {
	RoomID   string    `json:"room_id"`
	UserID   string    `json:"user_id"`
	JoinedAt time.Time `json:"joined_at"`
	Role     string    `json:"role"`
}

type UserLeftRoom struct {
	RoomID string    `json:"room_id"`
	UserID string    `json:"user_id"`
	LeftAt time.Time `json:"left_at"`
}
