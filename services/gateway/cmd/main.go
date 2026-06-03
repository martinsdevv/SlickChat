package main

import (
	"net/http"
	"os"

	"github.com/martinsdevv/slickchat/infrastructure/config"
	kafkainfra "github.com/martinsdevv/slickchat/infrastructure/kafka"
	"github.com/martinsdevv/slickchat/infrastructure/log"
	redisinfra "github.com/martinsdevv/slickchat/infrastructure/redis"
	"github.com/martinsdevv/slickchat/services/gateway/internal/ws"
)

func main() {
	rdb := redisinfra.NewClient()
	producer := kafkainfra.NewProducer(config.KafkaBroker())
	ticketStore := redisinfra.NewWSTicketStore(rdb)

	http.HandleFunc("/socket", ws.HandleWS(rdb, producer, ticketStore))

	log.Logger.Info("Gateway rodando em :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Logger.Error("server failed", "error", err)
		os.Exit(1)
	}
}
