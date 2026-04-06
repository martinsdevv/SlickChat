package main

import (
	"os"

	kafkainfra "github.com/martinsdevv/slickchat/infrastructure/kafka"
	"github.com/martinsdevv/slickchat/infrastructure/log"
	"github.com/martinsdevv/slickchat/infrastructure/postgres"
	"github.com/martinsdevv/slickchat/services/workers/persistence"
)

func main() {
	dsn := "postgres://postgres:postgres@localhost:5432/slickchat?sslmode=disable"

	db, err := postgres.NewConnection(dsn)
	if err != nil {
		log.Logger.Error("failed to connect to postgres", err)
		os.Exit(1)
	}

	repo := postgres.NewMessageRepository(db)

	handler := persistence.NewHandler(repo)
	log.Logger.Info("starting persistence worker")

	kafkainfra.StartConsumer(
		"localhost:9092",
		"message-events",
		"persistence-group",
		handler.Handle,
	)
}
