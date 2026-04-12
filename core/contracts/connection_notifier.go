package contracts

import (
	"context"

	"github.com/google/uuid"
)

// ConnectionNotifier publishes a force-disconnect signal to all active WebSocket
// connections belonging to a user. The gateway listens on the same channel and
// closes the connection cleanly.
type ConnectionNotifier interface {
	ForceDisconnectUser(ctx context.Context, userID uuid.UUID) error
}
