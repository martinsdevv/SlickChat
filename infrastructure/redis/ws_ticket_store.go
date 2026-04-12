package redis

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/martinsdevv/slickchat/core/contracts"
	"github.com/martinsdevv/slickchat/core/domain"
	"github.com/redis/go-redis/v9"
)

const wsTicketPrefix = "ws_ticket:"

type WSTicketStore struct {
	rdb *redis.Client
}

func NewWSTicketStore(rdb *redis.Client) contracts.WSTicketStore {
	return &WSTicketStore{rdb: rdb}
}

func (s *WSTicketStore) Save(ctx context.Context, ticketHash string, userID uuid.UUID, ttl time.Duration) error {
	return s.rdb.Set(ctx, wsTicketPrefix+ticketHash, userID.String(), ttl).Err()
}

// GetAndDelete atomically retrieves and removes the ticket from Redis.
// Returns domain.ErrSessionNotFound if the ticket does not exist or has expired.
func (s *WSTicketStore) GetAndDelete(ctx context.Context, ticketHash string) (uuid.UUID, error) {
	val, err := s.rdb.GetDel(ctx, wsTicketPrefix+ticketHash).Result()
	if err == redis.Nil {
		return uuid.Nil, domain.ErrSessionNotFound
	}
	if err != nil {
		return uuid.Nil, err
	}

	userID, err := uuid.Parse(val)
	if err != nil {
		return uuid.Nil, domain.ErrSessionNotFound
	}

	return userID, nil
}
