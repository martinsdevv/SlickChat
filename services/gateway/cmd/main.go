package main

import (
	"net/http"
	"os"

	kafkainfra "github.com/martinsdevv/slickchat/infrastructure/kafka"
	"github.com/martinsdevv/slickchat/infrastructure/log"
	redisinfra "github.com/martinsdevv/slickchat/infrastructure/redis"
	"github.com/martinsdevv/slickchat/services/gateway/internal/ws"
)

func main() {
	log.Init()
	rdb := redisinfra.NewClient()
	producer := kafkainfra.NewProducer("localhost:9092")

	http.HandleFunc("/socket", ws.HandleWS(rdb, producer))

	log.Logger.Info("Gateway rodando em :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Logger.Error("server failed", "error", err)
		os.Exit(1)
	}
}
