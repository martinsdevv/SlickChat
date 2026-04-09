package ws

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
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

type roomContextResponse struct {
	RoomID       string     `json:"room_id"`
	Type         string     `json:"type"`
	TTL          int        `json:"ttl"`
	ParanoidMode bool       `json:"paranoid_mode"`
	ZeroLogging  bool       `json:"zero_logging"`
	ExpiresAt    *time.Time `json:"expires_at,omitempty"`
	Role         string     `json:"role"`
}

func fetchRoomContext(ctx context.Context, apiBaseURL string, roomID string, userID string) (*domain.Room, *domain.RoomMembership, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		apiBaseURL+"/room-context?room_id="+roomID+"&user_id="+userID,
		nil,
	)
	if err != nil {
		return nil, nil, err
	}

	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusNotFound {
			return nil, nil, errors.New("not_in_room")
		}
		return nil, nil, errors.New("room_context_failed")
	}

	var rc roomContextResponse
	if err := json.NewDecoder(resp.Body).Decode(&rc); err != nil {
		return nil, nil, err
	}

	roomUUID, err := uuid.Parse(rc.RoomID)
	if err != nil {
		return nil, nil, err
	}

	rt := domain.RoomType(rc.Type)
	switch rt {
	case domain.RoomTypePublic, domain.RoomTypePrivate, domain.RoomTypeDirect, domain.RoomTypeTemporary:
		// ok
	default:
		return nil, nil, errors.New("invalid_room_type")
	}

	role := domain.Role(rc.Role)
	switch role {
	case domain.RoleAdmin, domain.RoleModerator, domain.RoleMember:
		// ok
	default:
		return nil, nil, errors.New("invalid_role")
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, nil, err
	}

	room := &domain.Room{
		ID:           roomUUID,
		Type:         rt,
		TTL:          rc.TTL,
		ParanoidMode: rc.ParanoidMode,
		ZeroLogging:  rc.ZeroLogging,
		ExpiresAt:    rc.ExpiresAt,
	}

	membership := &domain.RoomMembership{
		UserID: userUUID,
		Role:   role,
	}

	return room, membership, nil
}

func handleSendMessage(rdb *redis.Client, producer *kafkainfra.Producer, client *Client, userID string, payload SendMessagePayload) {
	ctx := context.Background()

	room, membership, err := fetchRoomContext(ctx, "http://localhost:8081", payload.RoomID, userID)
	if err != nil {
		if err.Error() == "not_in_room" {
			sendError(client, "not_in_room")
			return
		}
		log.Logger.Error("room context failed", "error", err)
		sendError(client, "internal_error")
		return
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

	case events.EventTypeMessageExpired:
		sendToConnection(connectionID, "message.expired", json.RawMessage(event.Payload))
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
	if payload.MessageID == "" {
		sendError(client, "invalid_message_id")
		return
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		sendError(client, "invalid_user_id")
		return
	}

	ctx := context.Background()
	room, membership, err := fetchRoomContext(ctx, "http://localhost:8081", payload.RoomID, userID)
	if err != nil {
		if err.Error() == "not_in_room" {
			sendError(client, "not_in_room")
			return
		}
		log.Logger.Error("room context failed", "error", err)
		sendError(client, "internal_error")
		return
	}

	messageUUID, err := uuid.Parse(payload.MessageID)
	if err != nil {
		sendError(client, "invalid_message_id")
		return
	}

	err = application.DeleteMessage(
		producer,
		room,
		membership,
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
