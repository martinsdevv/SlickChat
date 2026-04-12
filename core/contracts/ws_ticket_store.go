package contracts

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type WSTicketStore interface {
	Save(ctx context.Context, ticketHash string, userID uuid.UUID, ttl time.Duration) error
	GetAndDelete(ctx context.Context, ticketHash string) (uuid.UUID, error)
}
