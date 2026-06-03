package contracts

import (
	"context"

	"github.com/google/uuid"
	"github.com/martinsdevv/slickchat/core/domain"
)

type RoomRepository interface {
	Save(ctx context.Context, room *domain.Room) error
	GetByID(ctx context.Context, roomID uuid.UUID) (*domain.Room, error)
	ListPublic(ctx context.Context, limit int) ([]*domain.Room, error)
	ListByUser(ctx context.Context, userID uuid.UUID, limit int) ([]*domain.Room, error)
	SetAvatarObjectKey(ctx context.Context, roomID uuid.UUID, objectKey string) (previousKey string, err error)
	SetBannerObjectKey(ctx context.Context, roomID uuid.UUID, objectKey string) (previousKey string, err error)
}

// MemberInfo is a read model combining membership and user identity.
type MemberInfo struct {
	UserID          uuid.UUID
	Handle          string
	Role            domain.Role
	AvatarObjectKey string
}

type RoomMembershipRepository interface {
	Get(ctx context.Context, roomID uuid.UUID, userID uuid.UUID) (*domain.RoomMembership, error)
	Add(ctx context.Context, roomID uuid.UUID, userID uuid.UUID, role domain.Role) error
	Remove(ctx context.Context, roomID uuid.UUID, userID uuid.UUID) error
	ListByRoom(ctx context.Context, roomID uuid.UUID) ([]MemberInfo, error)
}
