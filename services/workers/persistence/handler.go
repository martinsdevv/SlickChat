package persistence

import (
	"context"
	"encoding/json"

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

	msg := &domain.Message{
		ID:               payload.MessageID,
		RoomID:           payload.RoomID,
		SenderID:         payload.SenderID,
		Content:          payload.Content,
		MessageType:      payload.MessageType,
		TTL:              payload.TTL,
		CreatedAt:        payload.Timestamp,
		ExpiresAt:        payload.ExpiresAt,
		DestroyAfterRead: payload.DestroyAfterRead,
	}

	if err := h.repo.Save(context.Background(), msg); err != nil {
		log.Logger.Error("failed to persist message", err, "message_id:", msg.ID)
		return
	}

	log.Logger.Info("message persisted", "message_id:", msg.ID)
}
