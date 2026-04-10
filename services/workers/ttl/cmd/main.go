package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	kafkainfra "github.com/martinsdevv/slickchat/infrastructure/kafka"
	"github.com/martinsdevv/slickchat/infrastructure/log"
	"github.com/martinsdevv/slickchat/infrastructure/postgres"
	"github.com/martinsdevv/slickchat/services/workers/ttl"
)

func main() {
	dsn := "postgres://postgres:postgres@localhost:5432/slickchat?sslmode=disable"

	db, err := postgres.NewConnection(dsn)
	if err != nil {
		log.Logger.Error("failed to connect to postgres", "error", err)
		os.Exit(1)
	}

	repo := postgres.NewMessageRepository(db)
	producer := kafkainfra.NewProducer("localhost:9092")

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	log.Logger.Info("starting ttl worker")

	ttl.Run(ctx, repo, producer, 5*time.Second, 100)
}
