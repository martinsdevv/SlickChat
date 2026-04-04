package main

import (
	"log"
	"net/http"

	"github.com/martinsdevv/slickchat/services/gateway/internal/redis"
	"github.com/martinsdevv/slickchat/services/gateway/internal/ws"
)

func main() {
	rdb := redis.NewClient()

	go ws.StartSubscriber(rdb)

	http.HandleFunc("/socket", ws.HandleWS(rdb))

	log.Println("Gateway rodando em :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
