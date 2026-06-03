package domain

import (
	"time"

	"github.com/google/uuid"
)

type MediaType string

const (
	MediaTypeImage MediaType = "IMAGE"
	MediaTypeVideo MediaType = "VIDEO"
	MediaTypeFile  MediaType = "FILE"
	MediaTypeAudio MediaType = "AUDIO"
)

type Attachment struct {
	ID        uuid.UUID
	MessageID uuid.UUID
	RoomID    uuid.UUID
	ObjectKey string
	Caption   string
	MediaType MediaType
	SizeBytes int64
	CreatedAt time.Time
}
