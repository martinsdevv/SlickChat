package main

import (
	"github.com/martinsdevv/slickchat/infrastructure/log"
	redisinfra "github.com/martinsdevv/slickchat/infrastructure/redis"
	fanout "github.com/martinsdevv/slickchat/services/workers/fanout/internal"
)

func main() {
	log.Init()

	rdb := redisinfra.NewClient()

	log.Logger.Info("Fanout worker rodando")

	fanout.StartConsumer("localhost:9092", fanout.FanoutHandler(rdb))
}
