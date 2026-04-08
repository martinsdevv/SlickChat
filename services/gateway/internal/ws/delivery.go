package ws

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/martinsdevv/slickchat/core/application"
	"github.com/martinsdevv/slickchat/core/domain"
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

	room := &domain.Room{
		ID:          uuid.MustParse(payload.RoomID),
		ZeroLogging: false,
	}

	membership := &domain.RoomMembership{
		UserID: uuid.MustParse(userID),
		Role:   domain.RoleMember,
	}

	messageID, err := application.SendMessage(
		producer,
		room,
		membership,
		uuid.MustParse(userID),
		payload.Content,
	)

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
	switch event.EventType {

	case events.EventTypeMessageSent:
		sendToConnection(connectionID, "message.received", json.RawMessage(event.Payload))

	case events.EventTypeMessageDelivered:
		sendToConnection(connectionID, "message.delivered", json.RawMessage(event.Payload))

	case events.EventTypeMessageRead:
		sendToConnection(connectionID, "message.read", json.RawMessage(event.Payload))

	case events.EventTypeMessageDeleted:
		sendToConnection(connectionID, "message.deleted", json.RawMessage(event.Payload))
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

func handleDeleteMessage(
	rdb *redis.Client,
	producer *kafkainfra.Producer,
	client *Client,
	userID string,
	payload MessageDeletePayload,
) {
	if !isUserInRoom(rdb, userID, payload.RoomID) {
		sendError(client, "not_in_room")
		return
	}

	if payload.MessageID == "" {
		sendError(client, "invalid_message_id")
		return
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		sendError(client, "invalid_user_id")
		return
	}

	roomUUID, err := uuid.Parse(payload.RoomID)
	if err != nil {
		sendError(client, "invalid_room_id")
		return
	}

	messageUUID, err := uuid.Parse(payload.MessageID)
	if err != nil {
		sendError(client, "invalid_message_id")
		return
	}

	err = application.DeleteMessage(
		producer,
		&domain.Room{ID: roomUUID},
		&domain.RoomMembership{
			UserID: userUUID,
			Role:   domain.RoleMember,
		},
		userUUID,
		messageUUID,
		userUUID, // TEMP
	)

	if err != nil {
		sendError(client, "delete_failed")
		return
	}

	sendAck(client)
}
