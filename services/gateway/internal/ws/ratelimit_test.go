package ws

import (
	"context"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func redisClientOrSkip(t *testing.T) *redis.Client {
	t.Helper()
	addr := os.Getenv("REDIS_TEST_ADDR")
	if addr == "" {
		addr = "127.0.0.1:6379"
	}
	rdb := redis.NewClient(&redis.Options{Addr: addr})
	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		t.Skipf("redis indisponível em %s: %v", addr, err)
	}
	t.Cleanup(func() { _ = rdb.Close() })
	return rdb
}

func TestWsRateAllowed_withinLimit(t *testing.T) {
	rdb := redisClientOrSkip(t)
	userID := uuid.New().String()
	action := "send_test_" + uuid.New().String()
	const max = 5

	for i := 0; i < max; i++ {
		assert.True(t, wsRateAllowed(rdb, userID, action, max), "iteração %d", i)
	}
	assert.False(t, wsRateAllowed(rdb, userID, action, max), "deve bloquear após exceder o limite")
}

func TestWsRateAllowed_separateKeysPerUser(t *testing.T) {
	rdb := redisClientOrSkip(t)
	action := "send_test_" + uuid.New().String()
	u1 := uuid.New().String()
	u2 := uuid.New().String()
	const max = 2

	assert.True(t, wsRateAllowed(rdb, u1, action, max))
	assert.True(t, wsRateAllowed(rdb, u1, action, max))
	assert.False(t, wsRateAllowed(rdb, u1, action, max))

	assert.True(t, wsRateAllowed(rdb, u2, action, max), "outro user_id não deve compartilhar contador")
	require.NoError(t, rdb.Del(context.Background(), "ratelimit:ws:"+u1+":"+action, "ratelimit:ws:"+u2+":"+action).Err())
}
