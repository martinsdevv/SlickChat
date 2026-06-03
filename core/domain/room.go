package domain

import (
	"time"

	"github.com/google/uuid"
)

type RoomType string

const (
	RoomTypePublic    RoomType = "PUBLIC"
	RoomTypePrivate   RoomType = "PRIVATE"
	RoomTypeDirect    RoomType = "DIRECT"
	RoomTypeTemporary RoomType = "TEMPORARY"
)

type Room struct {
	ID              uuid.UUID
	Name            string
	Description     string
	Type            RoomType
	OwnerID         uuid.UUID
	TTL             int
	ParanoidMode    bool
	ZeroLogging     bool
	AvatarObjectKey string
	BannerObjectKey string
	CreatedAt       time.Time
	ExpiresAt       *time.Time
}

type RoomMembership struct {
	UserID uuid.UUID
	Role   Role
}

func (r *Room) IsExpired(now time.Time) bool {
	if r.ExpiresAt == nil {
		return false
	}
	return now.After(*r.ExpiresAt)
}

func (r *Room) CanPersistMessages() bool {
	if r.ZeroLogging {
		return false
	}
	if r.ParanoidMode {
		return false
	}
	return true
}

// CanPersistAttachments follows message persistence: ephemeral rooms must not keep files after destroy.
func (r *Room) CanPersistAttachments() bool {
	return r.CanPersistMessages()
}
func (r *Room) CanUserSendMessage(userID uuid.UUID, role Role, now time.Time) error {
	if r.IsExpired(now) {
		return ErrRoomExpired
	}

	// por enquanto: qualquer membro pode enviar
	// (regras mais complexas tipo mute vêm depois)
	return nil
}

func (r *Room) CanDeleteMessage(
	userID uuid.UUID,
	membership *RoomMembership,
	messageSenderID uuid.UUID,
	now time.Time,
) error {
	if r.IsExpired(now) {
		return ErrRoomExpired
	}

	// dono da sala pode deletar
	if r.OwnerID != uuid.Nil && userID == r.OwnerID {
		return nil
	}

	// dono da mensagem pode deletar
	if userID == messageSenderID {
		return nil
	}

	// ou moderador/admin
	if membership != nil && membership.CanModerate() {
		return nil
	}

	return ErrPermissionDenied
}
func (m *RoomMembership) CanModerate() bool {
	return m.Role == RoleAdmin || m.Role == RoleModerator
}
