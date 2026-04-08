package contracts

import (
	"context"

	"github.com/google/uuid"
	"github.com/martinsdevv/slickchat/core/domain"
)

type MessageRepository interface {
	Save(ctx context.Context, msg *domain.Message) error
	ListByRoom(ctx context.Context, roomID uuid.UUID, limit int) ([]*domain.Message, error)
	Delete(ctx context.Context, messageID uuid.UUID) error
}
