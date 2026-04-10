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
	OwnerID      string     `json:"owner_id,omitempty"`
	TTL          int        `json:"ttl"`
	ParanoidMode bool       `json:"paranoid_mode"`
	ZeroLogging  bool       `json:"zero_logging"`
	ExpiresAt    *time.Time `json:"expires_at,omitempty"`
	Role         string     `json:"role"`
}

type messageContextResponse struct {
	MessageID        string     `json:"message_id"`
	RoomID           string     `json:"room_id"`
	SenderID         string     `json:"sender_id"`
	DestroyAfterRead bool       `json:"destroy_after_read"`
	ExpiresAt        *time.Time `json:"expires_at,omitempty"`
}

// getHTTP faz GET com retry em falha de rede ou 5xx/429 (API local pode oscilar sob carga).
func getHTTP(ctx context.Context, url string) (*http.Response, error) {
	client := &http.Client{Timeout: 5 * time.Second}
	var lastErr error
	backoff := 80 * time.Millisecond
	for attempt := 0; attempt < 3; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff):
			}
			backoff *= 2
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return nil, err
		}
		resp, err := client.Do(req)
		if err != nil {
			lastErr = err
			continue
		}
		if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
			resp.Body.Close()
			lastErr = errors.New("retryable http status")
			continue
		}
		return resp, nil
	}
	if lastErr == nil {
		lastErr = errors.New("http get failed")
	}
	return nil, lastErr
}

const roomContextCacheTTL = 30 * time.Second

func roomContextCacheKey(roomID, userID string) string {
	return "room_context:" + roomID + ":" + userID
}

func roomFromContextResponse(rc roomContextResponse, userID string) (*domain.Room, *domain.RoomMembership, error) {
	roomUUID, err := uuid.Parse(rc.RoomID)
	if err != nil {
		return nil, nil, err
	}

	rt := domain.RoomType(rc.Type)
	switch rt {
	case domain.RoomTypePublic, domain.RoomTypePrivate, domain.RoomTypeDirect, domain.RoomTypeTemporary:
	default:
		return nil, nil, errors.New("invalid_room_type")
	}

	role := domain.Role(rc.Role)
	switch role {
	case domain.RoleAdmin, domain.RoleModerator, domain.RoleMember:
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
	if rc.OwnerID != "" {
		ownerUUID, err := uuid.Parse(rc.OwnerID)
		if err != nil {
			return nil, nil, err
		}
		room.OwnerID = ownerUUID
	}

	membership := &domain.RoomMembership{
		UserID: userUUID,
		Role:   role,
	}

	return room, membership, nil
}

func fetchRoomContext(ctx context.Context, rdb *redis.Client, apiBaseURL string, roomID string, userID string) (*domain.Room, *domain.RoomMembership, error) {
	cacheKey := roomContextCacheKey(roomID, userID)
	if rdb != nil {
		if b, err := rdb.Get(ctx, cacheKey).Bytes(); err == nil {
			var rc roomContextResponse
			if json.Unmarshal(b, &rc) == nil {
				return roomFromContextResponse(rc, userID)
			}
		}
	}

	resp, err := getHTTP(ctx, apiBaseURL+"/room-context?room_id="+roomID+"&user_id="+userID)
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

	if rdb != nil {
		if raw, err := json.Marshal(rc); err == nil {
			_ = rdb.Set(ctx, cacheKey, raw, roomContextCacheTTL).Err()
		}
	}

	return roomFromContextResponse(rc, userID)
}

func fetchMessageContext(ctx context.Context, apiBaseURL string, messageID string) (*messageContextResponse, error) {
	resp, err := getHTTP(ctx, apiBaseURL+"/message-context?message_id="+messageID)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("message_context_failed")
	}

	var mc messageContextResponse
	if err := json.NewDecoder(resp.Body).Decode(&mc); err != nil {
		return nil, err
	}

	return &mc, nil
}

func handleSendMessage(rdb *redis.Client, producer *kafkainfra.Producer, client *Client, userID string, payload SendMessagePayload) {
	ctx := context.Background()

	if payload.RoomID == "" || payload.Content == "" {
		sendError(client, "invalid_payload")
		return
	}
	if _, err := uuid.Parse(payload.RoomID); err != nil {
		sendError(client, "invalid_room_id")
		return
	}
	if len(payload.Content) > 2000 {
		sendError(client, "content_too_long")
		return
	}

	if !wsRateAllowed(rdb, userID, "send", 45) {
		sendError(client, "rate_limited")
		return
	}

	room, membership, err := fetchRoomContext(ctx, rdb, "http://localhost:8081", payload.RoomID, userID)
	if err != nil {
		if err.Error() == "not_in_room" {
			sendError(client, "not_in_room")
			return
		}
		log.Logger.Error("room context failed", "error", err)
		sendError(client, "internal_error")
		return
	}

	messageID := uuid.New()
	msgKey := "message:" + messageID.String()
	if err := rdb.HSet(ctx, msgKey, map[string]interface{}{
		"sender_id": userID,
	}).Err(); err != nil {
		log.Logger.Error("redis hset message", "error", err)
		sendError(client, "internal_error")
		return
	}
	_ = rdb.Expire(ctx, msgKey, time.Hour*24).Err()

	if err := application.SendMessageWithID(
		producer,
		room,
		membership,
		uuid.MustParse(userID),
		messageID,
		payload.Content,
	); err != nil {
		_ = rdb.Del(ctx, msgKey).Err()
		log.Logger.Error("Erro ao enviar mensagem", "error", err)
		sendError(client, "internal_error")
		return
	}

	sendAck(client)
}

func handleMessageDelivered(
	rdb *redis.Client,
	producer *kafkainfra.Producer,
	client *Client,
	userID string,
	payload MessageDeliveredPayload,
) {
	if !wsRateAllowed(rdb, userID, "delivered", 120) {
		sendError(client, "rate_limited")
		return
	}

	if !isUserInRoom(rdb, userID, payload.RoomID) {
		sendError(client, "not_in_room")
		return
	}

	ctx := context.Background()
	senderID, _ := rdb.HGet(ctx, "message:"+payload.MessageID, "sender_id").Result()

	application.DeliverMessage(producer, userID, payload.RoomID, payload.MessageID, senderID)
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
	if !wsRateAllowed(rdb, userID, "read", 120) {
		sendError(client, "rate_limited")
		return
	}

	if !isUserInRoom(rdb, userID, payload.RoomID) {
		sendError(client, "not_in_room")
		return
	}

	if _, err := uuid.Parse(payload.RoomID); err != nil {
		sendError(client, "invalid_room_id")
		return
	}
	if _, err := uuid.Parse(payload.MessageID); err != nil {
		sendError(client, "invalid_message_id")
		return
	}

	ctx := context.Background()
	senderID, _ := rdb.HGet(ctx, "message:"+payload.MessageID, "sender_id").Result()

	_ = application.ReadMessage(producer, userID, payload.RoomID, payload.MessageID, senderID)

	mc, err := fetchMessageContext(ctx, "http://localhost:8081", payload.MessageID)
	if err != nil {
		// read já foi publicado; falha de auto-delete não pode derrubar a sessão
		log.Logger.Error("message context failed", "error", err)
		return
	}
	if mc.RoomID != payload.RoomID {
		return
	}
	if mc.DestroyAfterRead {
		_ = application.AutoDeleteMessage(producer, payload.RoomID, payload.MessageID)
	}
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

	if !wsRateAllowed(rdb, userID, "delete", 25) {
		sendError(client, "rate_limited")
		return
	}

	if _, err := uuid.Parse(payload.RoomID); err != nil {
		sendError(client, "invalid_room_id")
		return
	}
	if _, err := uuid.Parse(payload.MessageID); err != nil {
		sendError(client, "invalid_message_id")
		return
	}

	room, membership, err := fetchRoomContext(ctx, rdb, "http://localhost:8081", payload.RoomID, userID)
	if err != nil {
		if err.Error() == "not_in_room" {
			sendError(client, "not_in_room")
			return
		}
		log.Logger.Error("room context failed", "error", err)
		sendError(client, "internal_error")
		return
	}

	mc, err := fetchMessageContext(ctx, "http://localhost:8081", payload.MessageID)
	if err != nil {
		log.Logger.Error("message context failed", "error", err)
		sendError(client, "delete_failed")
		return
	}

	if mc.RoomID != payload.RoomID {
		sendError(client, "delete_failed")
		return
	}

	messageUUID := uuid.MustParse(payload.MessageID)
	senderUUID, err := uuid.Parse(mc.SenderID)
	if err != nil {
		sendError(client, "delete_failed")
		return
	}

	err = application.DeleteMessage(
		producer,
		room,
		membership,
		userUUID,
		messageUUID,
		senderUUID,
	)

	if err != nil {
		sendError(client, "delete_failed")
		return
	}

	sendAck(client)
}
