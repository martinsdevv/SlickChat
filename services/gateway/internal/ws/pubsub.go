package ws

import (
	"encoding/json"

	redisclient "github.com/martinsdevv/slickchat/services/gateway/internal/redis"
	goredis "github.com/redis/go-redis/v9"
)

func StartSubscriber(rdb *goredis.Client) {
	pubsub := rdb.PSubscribe(redisclient.Ctx, "room:*")

	ch := pubsub.Channel()

	for msg := range ch {
		var event MessageSent
		json.Unmarshal([]byte(msg.Payload), &event)

		sendToRoom(rdb, event.RoomID, event)
	}
}

func sendToUser(rdb *goredis.Client, userID string, event MessageSent) {
	connIDs, _ := rdb.SMembers(redisclient.Ctx, "user_connections:"+userID).Result()

	for _, connID := range connIDs {
		sendToConnection(connID, event)
	}
}

func sendToConnection(connectionID string, event MessageSent) {
	mu.Lock()
	client, ok := clients[connectionID]
	mu.Unlock()

	if !ok {
		return
	}

	client.Write(map[string]interface{}{
		"type":    "message_received",
		"payload": event,
	})
}

func sendToRoom(rdb *goredis.Client, roomID string, event MessageSent) {
	userIDs, _ := rdb.SMembers(redisclient.Ctx, "room_members:"+roomID).Result()

	for _, userID := range userIDs {
		sendToUser(rdb, userID, event)
	}
}
