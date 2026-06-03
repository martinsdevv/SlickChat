package contracts

import (
	"context"

	"github.com/google/uuid"
	"github.com/martinsdevv/slickchat/core/domain"
)

type AttachmentRepository interface {
	Save(ctx context.Context, attachment *domain.Attachment) error
	ReplaceForMessage(ctx context.Context, attachment *domain.Attachment) error
	ListByMessageID(ctx context.Context, messageID uuid.UUID) ([]*domain.Attachment, error)
	DeleteByMessageID(ctx context.Context, messageID uuid.UUID) ([]string, error)
}
