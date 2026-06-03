package main

import (
	"github.com/martinsdevv/slickchat/infrastructure/config"
	kafkainfra "github.com/martinsdevv/slickchat/infrastructure/kafka"
	"github.com/martinsdevv/slickchat/infrastructure/log"
	"github.com/martinsdevv/slickchat/infrastructure/media"
	"github.com/martinsdevv/slickchat/infrastructure/postgres"
	redisinfra "github.com/martinsdevv/slickchat/infrastructure/redis"
	"github.com/martinsdevv/slickchat/services/workers/fanout"
)

func main() {
	rdb := redisinfra.NewClient()
	dsn := config.LoadDBConfig()
	db, err := postgres.NewConnection(dsn.PGURL())
	if err != nil {
		panic(err)
	}
	membershipRepo := postgres.NewRoomMembershipRepository(db)

	log.Logger.Info("starting fanout worker")

	objectStorage := media.NewObjectStorageFromConfig(config.LoadMediaConfig())

	kafkainfra.StartConsumer(
		config.KafkaBroker(),
		"message-events",
		"fanout-group",
		fanout.FanoutHandler(rdb, membershipRepo, objectStorage),
	)
}
