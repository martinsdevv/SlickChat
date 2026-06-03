package redis

import (
	"context"

	"github.com/martinsdevv/slickchat/infrastructure/config"
	"github.com/redis/go-redis/v9"
)

var Ctx = context.Background()

func NewClient() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr: config.RedisAddr(),
	})
}
