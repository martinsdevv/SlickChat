package ws

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

const wsRateWindow = time.Minute

// wsRateAllowed retorna false se o usuário excedeu o limite na janela (INCR + EXPIRE).
func wsRateAllowed(rdb *redis.Client, userID, action string, maxPerWindow int) bool {
	ctx := context.Background()
	key := "ratelimit:ws:" + userID + ":" + action
	n, err := rdb.Incr(ctx, key).Result()
	if err != nil {
		return true
	}
	if n == 1 {
		_ = rdb.Expire(ctx, key, wsRateWindow).Err()
	}
	return n <= int64(maxPerWindow)
}
