package media

import (
	"strings"

	"github.com/google/uuid"
)

// RoomIDFromObjectKey extracts room UUID from keys like rooms/{id}/avatar/...
func RoomIDFromObjectKey(objectKey string) (uuid.UUID, bool) {
	parts := strings.Split(objectKey, "/")
	if len(parts) < 2 || parts[0] != "rooms" {
		return uuid.Nil, false
	}
	id, err := uuid.Parse(parts[1])
	if err != nil {
		return uuid.Nil, false
	}
	return id, true
}

// UserIDFromObjectKey extracts user UUID from keys like users/{id}/avatar/...
func UserIDFromObjectKey(objectKey string) (uuid.UUID, bool) {
	parts := strings.Split(objectKey, "/")
	if len(parts) < 2 || parts[0] != "users" {
		return uuid.Nil, false
	}
	id, err := uuid.Parse(parts[1])
	if err != nil {
		return uuid.Nil, false
	}
	return id, true
}

func IsRoomMediaKey(objectKey string) bool {
	if !strings.HasPrefix(objectKey, "rooms/") {
		return false
	}
	return strings.Contains(objectKey, "/avatar/") || strings.Contains(objectKey, "/banner/")
}

func IsUserAvatarKey(objectKey string) bool {
	return strings.HasPrefix(objectKey, "users/") && strings.Contains(objectKey, "/avatar/")
}

func IsMessageMediaKey(objectKey string) bool {
	return strings.HasPrefix(objectKey, "messages/")
}

func MessageRoomIDFromObjectKey(objectKey string) (uuid.UUID, bool) {
	parts := strings.Split(objectKey, "/")
	if len(parts) < 2 || parts[0] != "messages" {
		return uuid.Nil, false
	}
	id, err := uuid.Parse(parts[1])
	if err != nil {
		return uuid.Nil, false
	}
	return id, true
}
