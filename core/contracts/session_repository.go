package contracts

import (
	"context"

	"github.com/google/uuid"
	"github.com/martinsdevv/slickchat/core/domain"
)

type SessionRepository interface {
	Save(ctx context.Context, session *domain.Session) error
	GetByTokenHash(ctx context.Context, tokenHash string) (*domain.Session, error)
	DeleteByTokenHash(ctx context.Context, tokenHash string) error
	DeleteAllByUserID(ctx context.Context, userID uuid.UUID) error
}
