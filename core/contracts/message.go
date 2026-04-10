package contracts

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/martinsdevv/slickchat/core/domain"
)

type MessageRepository interface {
	// Save retorna linhas afetadas (0 = ON CONFLICT / duplicata, 1 = inserido).
	Save(ctx context.Context, msg *domain.Message) (int64, error)
	GetByID(ctx context.Context, messageID uuid.UUID) (*domain.Message, error)
	ListByRoom(ctx context.Context, roomID uuid.UUID, limit int) ([]*domain.Message, error)
	ListExpired(ctx context.Context, before time.Time, limit int) ([]*domain.Message, error)
	// Delete retorna linhas removidas (0 = já não existia, idempotente).
	Delete(ctx context.Context, messageID uuid.UUID) (int64, error)
}
