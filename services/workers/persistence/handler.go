package persistence

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/martinsdevv/slickchat/infrastructure/log"

	"github.com/martinsdevv/slickchat/core/contracts"
	"github.com/martinsdevv/slickchat/core/domain"
	"github.com/martinsdevv/slickchat/core/events"
)

type Handler struct {
	repo contracts.MessageRepository
}

func NewHandler(repo contracts.MessageRepository) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) Handle(event events.Event) {
	log.Logger.Info("processing event",
		"event_type", event.EventType,
	)

	switch event.EventType {

	case events.EventTypeMessageSent:
		h.handleMessageSent(event)

	case events.EventTypeMessageDeleted:
		h.handleMessageDeleted(event)

	case events.EventTypeMessageExpired:
		h.handleMessageExpired(event)

	default:
		log.Logger.Warn("unknown event type",
			"event_type", event.EventType,
		)
	}
}

func (h *Handler) handleMessageSent(event events.Event) {
	var payload events.MessageSent

	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		log.Logger.Error("failed to unmarshal payload", err)
		return
	}

	// regra mínima
	if payload.IsZeroLogging {
		log.Logger.Info("skipping zero logging message")
		return
	}

	id, err := uuid.Parse(payload.MessageID)
	if err != nil {
		log.Logger.Error("invalid message id", "error", err)
		return
	}

	roomID, err := uuid.Parse(payload.RoomID)
	if err != nil {
		log.Logger.Error("invalid room id", "error", err)
		return
	}

	senderID, err := uuid.Parse(payload.SenderID)
	if err != nil {
		log.Logger.Error("invalid sender id", "error", err)
		return
	}

	expiresAt := payload.ExpiresAt

	if expiresAt == nil && payload.TTL > 0 {
		t := payload.SentAt.Add(time.Duration(payload.TTL) * time.Second)
		expiresAt = &t
	}

	msg := &domain.Message{
		ID:               id,
		RoomID:           roomID,
		SenderID:         senderID,
		Content:          payload.Content,
		MessageType:      payload.MessageType,
		TTL:              payload.TTL,
		CreatedAt:        payload.SentAt,
		ExpiresAt:        expiresAt,
		DestroyAfterRead: payload.DestroyAfterRead,
	}

	if err := h.repo.Save(context.Background(), msg); err != nil {
		log.Logger.Error("failed to persist message", err, "message_id:", msg.ID)
		return
	}

	log.Logger.Info("message persisted", "message_id:", msg.ID)
}

func (h *Handler) handleMessageDeleted(event events.Event) {
	var payload events.MessageDeleted

	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		log.Logger.Error("failed to unmarshal payload", "error", err)
		return
	}

	id, err := uuid.Parse(payload.MessageID)
	if err != nil {
		log.Logger.Error("invalid message id", "error", err)
		return
	}

	if err := h.repo.Delete(context.Background(), id); err != nil {
		log.Logger.Error("failed to delete message", "error", err)
		return
	}

	log.Logger.Info("message deleted", "message_id", payload.MessageID)
}

func (h *Handler) handleMessageExpired(event events.Event) {
	var payload events.MessageExpired

	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		log.Logger.Error("failed to unmarshal payload", "error", err)
		return
	}

	id, err := uuid.Parse(payload.MessageID)
	if err != nil {
		log.Logger.Error("invalid message id", "error", err)
		return
	}

	if err := h.repo.Delete(context.Background(), id); err != nil {
		log.Logger.Error("failed to delete expired message", "error", err)
		return
	}

	log.Logger.Info("message expired removed", "message_id", payload.MessageID)
}
