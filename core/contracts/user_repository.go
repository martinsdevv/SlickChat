package contracts

import (
	"context"

	"github.com/google/uuid"
	"github.com/martinsdevv/slickchat/core/domain"
)

type UserRepository interface {
	Save(ctx context.Context, user *domain.User) error
	GetByHandle(ctx context.Context, username, discriminator string) (*domain.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	HandleExists(ctx context.Context, username, discriminator string) (bool, error)
}
