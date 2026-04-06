package domain

import "time"

type Message struct {
	ID               string
	RoomID           string
	SenderID         string
	Content          string
	MessageType      string
	TTL              int
	DestroyAfterRead bool
	CreatedAt        time.Time
	ExpiresAt        *time.Time
	IsZeroLogging    bool
}
