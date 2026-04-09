package contracts

import (
	"context"

	"github.com/google/uuid"
	"github.com/martinsdevv/slickchat/core/domain"
)

type RoomRepository interface {
	GetByID(ctx context.Context, roomID uuid.UUID) (*domain.Room, error)
}

type RoomMembershipRepository interface {
	Get(ctx context.Context, roomID uuid.UUID, userID uuid.UUID) (*domain.RoomMembership, error)
}
