package fanout

import (
	"context"
	"encoding/json"
	"time"

	"github.com/martinsdevv/slickchat/core/events"
	"github.com/martinsdevv/slickchat/infrastructure/log"
	"github.com/redis/go-redis/v9"
)

const fanoutDedupeTTL = 72 * time.Hour

// fanoutDedupe garante efeito colateral do fanout uma vez por (tipo, message_id) — replay Kafka não duplica WS/unread.
func fanoutDedupe(ctx context.Context, rdb *redis.Client, kind, messageID string) bool {
	if messageID == "" {
		return false
	}
	key := "fanout:dedupe:" + kind + ":" + messageID
	ok, err := rdb.SetNX(ctx, key, "1", fanoutDedupeTTL).Result()
	return err == nil && ok
}

type MessageDeliveredWithCount struct {
	events.MessageDelivered
	DeliveredCount int `json:"delivered_count"`
}

type MessageReadWithCount struct {
	events.MessageRead
	ReadCount int `json:"read_count"`
}

func FanoutHandler(rdb *redis.Client) func(events.Event) {
	return func(event events.Event) {
		switch event.EventType {

		case events.EventTypeMessageDelivered:
			handleMessageDelivered(event, rdb)

		case events.EventTypeMessageSent:
			handleMessageSent(event, rdb)

		case events.EventTypeMessageRead:
			handleMessageRead(event, rdb)

		case events.EventTypeMessageDeleted:
			handleMessageDeleted(event, rdb)

		case events.EventTypeMessageExpired:
			handleMessageExpired(event, rdb)
		}
	}
}

func handleMessageDelivered(event events.Event, rdb *redis.Client) {
	ctx := context.Background()

	var payload events.MessageDelivered
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		log.Logger.Warn("fanout poison payload", "event_type", event.EventType, "event_id", event.EventID, "error", err)
		return
	}

	senderID := getMessageSender(rdb, payload.MessageID)
	if senderID == "" {
		return
	}

	key := "msg:" + payload.MessageID + ":delivered"

	// idempotência
	added, _ := rdb.SAdd(ctx, key, payload.UserID).Result()
	if added == 0 {
		return // já processado
	}

	// TTL
	rdb.Expire(ctx, key, 24*time.Hour)

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
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		log.Logger.Warn("fanout poison payload", "event_type", event.EventType, "event_id", event.EventID, "error", err)
		return
	}

	if !fanoutDedupe(ctx, rdb, "sent", payload.MessageID) {
		return
	}

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

func decrUnreadClamp0(rdb *redis.Client, userID string, roomID string) {
	ctx := context.Background()

	key := "user:" + userID + ":room:" + roomID + ":unread"

	// decr e garante >= 0
	script := redis.NewScript(`
local v = redis.call("DECR", KEYS[1])
if v < 0 then
  redis.call("SET", KEYS[1], 0)
  return 0
end
return v
`)

	_, _ = script.Run(ctx, rdb, []string{key}).Result()
}

func handleMessageRead(event events.Event, rdb *redis.Client) {
	ctx := context.Background()

	var payload events.MessageRead
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		log.Logger.Warn("fanout poison payload", "event_type", event.EventType, "event_id", event.EventID, "error", err)
		return
	}

	senderID := getMessageSender(rdb, payload.MessageID)
	if senderID == "" {
		// Mensagem já removida do Redis (delete/expire) ou nunca existiu — ignora read fora de ordem.
		return
	}

	key := "msg:" + payload.MessageID + ":read"

	// idempotência
	added, _ := rdb.SAdd(ctx, key, payload.UserID).Result()
	if added == 0 {
		return
	}

	rdb.Expire(ctx, key, 24*time.Hour)

	// Uma mensagem lida: decrementa 1 (não apaga a chave — DEL deixava GET = nil).
	decrUnreadClamp0(rdb, payload.UserID, payload.RoomID)

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

func handleMessageDeleted(event events.Event, rdb *redis.Client) {
	ctx := context.Background()

	var payload events.MessageDeleted
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		log.Logger.Warn("fanout poison payload", "event_type", event.EventType, "event_id", event.EventID, "error", err)
		return
	}

	if !fanoutDedupe(ctx, rdb, "deleted", payload.MessageID) {
		return
	}

	members, _ := rdb.SMembers(ctx, "room_members:"+payload.RoomID).Result()
	senderID := getMessageSender(rdb, payload.MessageID)

	eventBytes, _ := json.Marshal(event)

	for _, userID := range members {
		if senderID != "" && userID != senderID {
			decrUnreadClamp0(rdb, userID, payload.RoomID)
		}

		connections, _ := rdb.SMembers(ctx, "user_connections:"+userID).Result()

		for _, connID := range connections {
			rdb.Publish(ctx, "connection:"+connID, eventBytes)
		}
	}

	rdb.Del(ctx, "message:"+payload.MessageID)
}

func handleMessageExpired(event events.Event, rdb *redis.Client) {
	ctx := context.Background()

	var payload events.MessageExpired
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		log.Logger.Warn("fanout poison payload", "event_type", event.EventType, "event_id", event.EventID, "error", err)
		return
	}

	if !fanoutDedupe(ctx, rdb, "expired", payload.MessageID) {
		return
	}

	members, _ := rdb.SMembers(ctx, "room_members:"+payload.RoomID).Result()
	senderID := getMessageSender(rdb, payload.MessageID)

	eventBytes, _ := json.Marshal(event)

	for _, userID := range members {
		if senderID != "" && userID != senderID {
			decrUnreadClamp0(rdb, userID, payload.RoomID)
		}

		connections, _ := rdb.SMembers(ctx, "user_connections:"+userID).Result()

		for _, connID := range connections {
			rdb.Publish(ctx, "connection:"+connID, eventBytes)
		}
	}

	rdb.Del(ctx, "message:"+payload.MessageID)
}
