package contracts

import (
	"context"

	"github.com/martinsdevv/slickchat/core/domain"
)

type MessageRepository interface {
	Save(ctx context.Context, msg *domain.Message) error
	ListByRoom(ctx context.Context, roomID string, limit int) ([]*domain.Message, error)
}
