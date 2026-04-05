package fanout

import (
	"context"
	"encoding/json"
	"time"

	"github.com/martinsdevv/slickchat/core/events"
	"github.com/redis/go-redis/v9"
)

type MessageDeliveredWithCount struct {
	events.MessageDelivered
	DeliveredCount int `json:"delivered_count"`
}

type MessageReadWithCount struct {
	events.MessageRead
	ReadCount int `json:"read_count"`
}

func handleMessageDelivered(event events.Event, rdb *redis.Client) {
	ctx := context.Background()

	var payload events.MessageDelivered
	json.Unmarshal(event.Payload, &payload)

	key := "msg:" + payload.MessageID + ":delivered"

	// idempotência
	added, _ := rdb.SAdd(ctx, key, payload.UserID).Result()
	if added == 0 {
		return // já processado
	}

	// TTL
	rdb.Expire(ctx, key, 24*time.Hour)

	senderID := getMessageSender(rdb, payload.MessageID)
	if senderID == "" {
		return
	}

	connections, _ := rdb.SMembers(ctx, "user_connections:"+senderID).Result()

	count, _ := rdb.SCard(ctx, key).Result()

	enriched := MessageDeliveredWithCount{
		MessageDelivered: payload,
		DeliveredCount:   int(count),
	}

	event.Payload, _ = json.Marshal(enriched)

	eventBytes, _ := json.Marshal(event)

	for _, connID := range connections {
		rdb.Publish(ctx, "connection:"+connID, eventBytes)
	}
}

func handleMessageSent(event events.Event, rdb *redis.Client) {
	ctx := context.Background()

	var payload events.MessageSent
	json.Unmarshal(event.Payload, &payload)

	rdb.Expire(ctx, "msg:"+payload.MessageID+":delivered", 24*time.Hour)
	rdb.Expire(ctx, "msg:"+payload.MessageID+":read", 24*time.Hour)

	members, _ := rdb.SMembers(ctx, "room_members:"+payload.RoomID).Result()

	eventBytes, _ := json.Marshal(event)

	for _, userID := range members {
		if userID != payload.SenderID {
			rdb.Incr(ctx, "user:"+userID+":room:"+payload.RoomID+":unread")
		}

		connections, _ := rdb.SMembers(ctx, "user_connections:"+userID).Result()

		for _, connID := range connections {
			rdb.Publish(ctx, "connection:"+connID, eventBytes)
		}
	}
}

func handleMessageRead(event events.Event, rdb *redis.Client) {
	ctx := context.Background()

	var payload events.MessageRead
	json.Unmarshal(event.Payload, &payload)

	key := "msg:" + payload.MessageID + ":read"

	// idempotência
	added, _ := rdb.SAdd(ctx, key, payload.UserID).Result()
	if added == 0 {
		return
	}

	rdb.Expire(ctx, key, 24*time.Hour)

	senderID := getMessageSender(rdb, payload.MessageID)
	if senderID == "" {
		return
	}

	rdb.Del(ctx, "user:"+payload.UserID+":room:"+payload.RoomID+":unread")

	connections, _ := rdb.SMembers(ctx, "user_connections:"+senderID).Result()

	count, _ := rdb.SCard(ctx, key).Result()
	enriched := MessageReadWithCount{
		MessageRead: payload,
		ReadCount:   int(count),
	}

	event.Payload, _ = json.Marshal(enriched)
	eventBytes, _ := json.Marshal(event)

	for _, connID := range connections {
		rdb.Publish(ctx, "connection:"+connID, eventBytes)
	}
}

func getMessageSender(rdb *redis.Client, messageID string) string {
	ctx := context.Background()

	val, err := rdb.HGet(ctx, "message:"+messageID, "sender_id").Result()
	if err != nil {
		return ""
	}

	return val
}

func isUserInRoom(rdb *redis.Client, userID, roomID string) bool {
	ctx := context.Background()

	exists, err := rdb.SIsMember(ctx, "room_members:"+roomID, userID).Result()
	if err != nil {
		return false
	}

	return exists
}
