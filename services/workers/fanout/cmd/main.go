package main

import (
	"context"
	"encoding/json"
	"log"

	"github.com/martinsdevv/slickchat/services/workers/fanout/internal"
	"github.com/redis/go-redis/v9"
)

func handleFanout(rdb *redis.Client, event fanout.MessageSent) {
	// Publicar no canal Redis para a sala
	payload, _ := json.Marshal(event)
	rdb.Publish(context.Background(), "room:"+event.RoomID, payload)
	log.Printf("Evento publicado para sala %s", event.RoomID)
}

func main() {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	fanout.StartConsumer("localhost:9092", func(event fanout.MessageSent) {
		handleFanout(rdb, event)
	})

	log.Println("Fanout worker rodando")
}
