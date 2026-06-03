package domain

import (
	"time"

	"github.com/google/uuid"
)

type Role string

const (
	RoleAdmin     Role = "ADMIN"
	RoleModerator Role = "MODERATOR"
	RoleMember    Role = "MEMBER"
)

type User struct {
	ID              uuid.UUID
	Username        string
	Discriminator   string
	PasswordHash    string
	RecoveryKeyHash string
	ParanoidMode    bool
	AvatarObjectKey string
	CreatedAt       time.Time
}

func (u *User) Handle() string {
	return u.Username + "#" + u.Discriminator
}
