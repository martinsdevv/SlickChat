package main

import (
	kafkainfra "github.com/martinsdevv/slickchat/infrastructure/kafka"
	"github.com/martinsdevv/slickchat/infrastructure/log"
	redisinfra "github.com/martinsdevv/slickchat/infrastructure/redis"
	"github.com/martinsdevv/slickchat/services/workers/fanout"
)

func main() {
	rdb := redisinfra.NewClient()

	log.Logger.Info("starting fanout worker")

	kafkainfra.StartConsumer(
		"localhost:9092",
		"message-events",
		"fanout-group",
		fanout.FanoutHandler(rdb),
	)
}
