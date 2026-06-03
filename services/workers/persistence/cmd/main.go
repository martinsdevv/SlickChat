package main

import (
	"os"

	kafkainfra "github.com/martinsdevv/slickchat/infrastructure/kafka"
	"github.com/martinsdevv/slickchat/infrastructure/log"
	"github.com/martinsdevv/slickchat/infrastructure/config"
	"github.com/martinsdevv/slickchat/infrastructure/media"
	"github.com/martinsdevv/slickchat/infrastructure/postgres"
	"github.com/martinsdevv/slickchat/services/workers/persistence"
)

func main() {
	db, err := postgres.NewConnection(config.LoadDBConfig().PGURL())
	if err != nil {
		log.Logger.Error("failed to connect to postgres", "error", err)
		os.Exit(1)
	}

	repo := postgres.NewMessageRepository(db)
	attachmentRepo := postgres.NewAttachmentRepository(db)

	objectStorage := media.NewObjectStorageFromConfig(config.LoadMediaConfig())
	handler := persistence.NewHandler(repo, attachmentRepo, objectStorage)
	log.Logger.Info("starting persistence worker")

	kafkainfra.StartConsumer(
		config.KafkaBroker(),
		"message-events",
		"persistence-group",
		handler.Handle,
	)
}
