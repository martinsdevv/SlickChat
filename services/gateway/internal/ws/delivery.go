package ws

import (
	"context"
	"encoding/json"
	"time"

	"github.com/martinsdevv/slickchat/core/application"
	"github.com/martinsdevv/slickchat/core/events"
	kafkainfra "github.com/martinsdevv/slickchat/infrastructure/kafka"
	"github.com/martinsdevv/slickchat/infrastructure/log"
	"github.com/redis/go-redis/v9"
)

type WSMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

func handleSendMessage(rdb *redis.Client, producer *kafkainfra.Producer, client *Client, userID string, payload SendMessagePayload) {
	ctx := context.Background()

	if !isUserInRoom(rdb, userID, payload.RoomID) {
		sendError(client, "not_in_room")
		return
	}

	messageID, err := application.SendMessage(producer, userID, payload.RoomID, payload.Content)
	if err != nil {
		log.Logger.Error("Erro ao enviar mensagem: ", err)
		sendError(client, "internal_error")
		return
	}

	rdb.HSet(ctx, "message:"+messageID, map[string]interface{}{
		"sender_id": userID,
	})

	rdb.Expire(ctx, "message:"+messageID, time.Hour*24)

	sendAck(client)
}

func handleMessageDelivered(
	rdb *redis.Client,
	producer *kafkainfra.Producer,
	client *Client,
	userID string,
	payload MessageDeliveredPayload,
) {
	if !isUserInRoom(rdb, userID, payload.RoomID) {
		sendError(client, "not_in_room")
		return
	}

	application.DeliverMessage(producer, userID, payload.RoomID, payload.MessageID)
}

func sendToConnection(connectionID, msgType string, payload interface{}) {
	mu.Lock()
	client, ok := clients[connectionID]
	mu.Unlock()

	if !ok {
		return
	}

	client.Write(WSMessage{
		Type:    msgType,
		Payload: payload,
	})
}

func handleIncomingEvent(connectionID string, event events.Event) {
	switch event.Type {

	case events.EventTypeMessageSent:
		sendToConnection(connectionID, "message.received", json.RawMessage(event.Payload))

	case events.EventTypeMessageDelivered:
		sendToConnection(connectionID, "message.delivered", json.RawMessage(event.Payload))

	case events.EventTypeMessageRead:
		sendToConnection(connectionID, "message.read", json.RawMessage(event.Payload))
	}
}

func handleMessageRead(
	rdb *redis.Client,
	producer *kafkainfra.Producer,
	client *Client,
	userID string,
	payload MessageReadPayload,
) {
	if !isUserInRoom(rdb, userID, payload.RoomID) {
		sendError(client, "not_in_room")
		return
	}

	application.ReadMessage(producer, userID, payload.RoomID, payload.MessageID)
}
