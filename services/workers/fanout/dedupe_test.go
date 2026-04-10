package fanout

import (
	"context"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func fanoutRedisOrSkip(t *testing.T) *redis.Client {
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

func TestFanoutDedupe_firstCallWins(t *testing.T) {
	rdb := fanoutRedisOrSkip(t)
	ctx := context.Background()
	mid := uuid.New().String()
	key := "fanout:dedupe:sent:" + mid
	t.Cleanup(func() { _ = rdb.Del(ctx, key).Err() })

	assert.True(t, fanoutDedupe(ctx, rdb, "sent", mid))
	assert.False(t, fanoutDedupe(ctx, rdb, "sent", mid), "replay Kafka / segunda chamada deve ser ignorada")
}

func TestFanoutDedupe_emptyMessageID(t *testing.T) {
	rdb := fanoutRedisOrSkip(t)
	ctx := context.Background()
	assert.False(t, fanoutDedupe(ctx, rdb, "sent", ""))
}

func TestFanoutDedupe_kindIsolated(t *testing.T) {
	rdb := fanoutRedisOrSkip(t)
	ctx := context.Background()
	mid := uuid.New().String()
	k1 := "fanout:dedupe:sent:" + mid
	k2 := "fanout:dedupe:deleted:" + mid
	t.Cleanup(func() { _ = rdb.Del(ctx, k1, k2).Err() })

	assert.True(t, fanoutDedupe(ctx, rdb, "sent", mid))
	assert.True(t, fanoutDedupe(ctx, rdb, "deleted", mid), "tipo diferente = chave Redis diferente")
}
