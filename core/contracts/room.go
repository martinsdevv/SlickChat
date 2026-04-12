package contracts

import (
	"context"

	"github.com/google/uuid"
	"github.com/martinsdevv/slickchat/core/domain"
)

type RoomRepository interface {
	Save(ctx context.Context, room *domain.Room) error
	GetByID(ctx context.Context, roomID uuid.UUID) (*domain.Room, error)
}

type RoomMembershipRepository interface {
	Get(ctx context.Context, roomID uuid.UUID, userID uuid.UUID) (*domain.RoomMembership, error)
	Add(ctx context.Context, roomID uuid.UUID, userID uuid.UUID, role domain.Role) error
	Remove(ctx context.Context, roomID uuid.UUID, userID uuid.UUID) error
}
