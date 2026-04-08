package domain

import (
	"time"

	"github.com/google/uuid"
)

type Message struct {
	ID               uuid.UUID
	RoomID           uuid.UUID
	SenderID         uuid.UUID
	Content          string
	MessageType      string
	TTL              int
	DestroyAfterRead bool
	CreatedAt        time.Time
	ExpiresAt        *time.Time
	IsZeroLogging    bool
}
