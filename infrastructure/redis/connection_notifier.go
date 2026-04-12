package redis

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/martinsdevv/slickchat/core/contracts"
	"github.com/martinsdevv/slickchat/core/events"
	"github.com/redis/go-redis/v9"
)

type ConnectionNotifier struct {
	rdb *redis.Client
}

func NewConnectionNotifier(rdb *redis.Client) contracts.ConnectionNotifier {
	return &ConnectionNotifier{rdb: rdb}
}

// ForceDisconnectUser publishes a user.logged_out.v1 event to every active
// WebSocket connection of the given user. The gateway's subscribeConnection
// goroutine receives the message and closes the connection.
func (n *ConnectionNotifier) ForceDisconnectUser(ctx context.Context, userID uuid.UUID) error {
	connIDs, err := n.rdb.SMembers(ctx, "user_connections:"+userID.String()).Result()
	if err != nil {
		return err
	}

	if len(connIDs) == 0 {
		return nil
	}

	payload := events.UserLoggedOut{
		UserID:   userID.String(),
		LoggedAt: time.Now().UTC(),
	}

	event, err := events.NewEvent(events.EventTypeUserLoggedOut, userID.String(), payload)
	if err != nil {
		return err
	}

	raw, err := json.Marshal(event)
	if err != nil {
		return err
	}

	for _, connID := range connIDs {
		_ = n.rdb.Publish(ctx, "connection:"+connID, raw).Err()
	}

	return nil
}
