package main

import (
	"context"
	"encoding/json"
	"fmt"

	fanout "github.com/martinsdevv/slickchat/services/workers/fanout/internal"
	"github.com/redis/go-redis/v9"
)

func main() {
	ctx := context.Background()

	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	fmt.Println("Fanout worker rodando")
	fanout.StartConsumer("localhost:9092", func(event fanout.MessageSent) {
		handleFanout(ctx, rdb, event)
	})
}

func handleFanout(ctx context.Context, rdb *redis.Client, event fanout.MessageSent) {
	userIDs, _ := rdb.SMembers(ctx, "room_members:"+event.RoomID).Result()

	data, _ := json.Marshal(event)

	if !isUserInRoom(rdb, event.SenderID, event.RoomID) {
		return
	}

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
