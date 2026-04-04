package main

import (
	"log"
	"net/http"

	kafkainfra "github.com/martinsdevv/slickchat/infrastructure/kafka"
	redisinfra "github.com/martinsdevv/slickchat/infrastructure/redis"
	"github.com/martinsdevv/slickchat/services/gateway/internal/ws"
)

func main() {
	rdb := redisinfra.NewClient()
	producer := kafkainfra.NewProducer("localhost:9092")

	http.HandleFunc("/socket", ws.HandleWS(rdb, producer))

	log.Println("Gateway rodando em :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
