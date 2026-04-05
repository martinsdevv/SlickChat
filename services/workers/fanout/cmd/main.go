package main

import (
	"context"
	"encoding/json"

	"github.com/martinsdevv/slickchat/core/events"
	"github.com/martinsdevv/slickchat/infrastructure/log"
	redisinfra "github.com/martinsdevv/slickchat/infrastructure/redis"
	fanout "github.com/martinsdevv/slickchat/services/workers/fanout/internal"
	"github.com/redis/go-redis/v9"
)

func main() {
	ctx := context.Background()
	log.Init()

	rdb := redisinfra.NewClient()

	log.Logger.Info("Fanout worker rodando")

	fanout.StartConsumer("localhost:9092", func(event events.Event) {

		switch event.Type {

		case events.EventTypeMessageSent:
			var payload events.MessageSent
			json.Unmarshal(event.Payload, &payload)

			handleFanout(ctx, rdb, payload)
		}
	})
}

func handleFanout(ctx context.Context, rdb *redis.Client, event events.MessageSent) {

	log.Logger.Info("fanout_start",
		"room_id", event.RoomID,
		"sender_id", event.SenderID,
	)

	if !isUserInRoom(rdb, event.SenderID, event.RoomID) {
		log.Logger.Error("user not in room")
		return
	}

	userIDs, _ := rdb.SMembers(ctx, "room_members:"+event.RoomID).Result()

	data, _ := json.Marshal(event)

	for _, userID := range userIDs {
		connIDs, _ := rdb.SMembers(ctx, "user_connections:"+userID).Result()

		for _, connID := range connIDs {
			rdb.Publish(ctx, "connection:"+connID, data)
		}
	}
}
func isUserInRoom(rdb *redis.Client, userID, roomID string) bool {
	ctx := context.Background()

	exists, err := rdb.SIsMember(ctx, "room_members:"+roomID, userID).Result()
	if err != nil {
		return false
	}

	return exists
}
