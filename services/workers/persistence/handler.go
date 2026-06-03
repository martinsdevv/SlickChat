package persistence

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/martinsdevv/slickchat/core/application"
	"github.com/martinsdevv/slickchat/core/contracts"
	"github.com/martinsdevv/slickchat/core/domain"
	"github.com/martinsdevv/slickchat/core/events"
	"github.com/martinsdevv/slickchat/infrastructure/log"
)

type Handler struct {
	repo        contracts.MessageRepository
	attachments contracts.AttachmentRepository
	storage     contracts.ObjectStorage
}

func NewHandler(
	repo contracts.MessageRepository,
	attachments contracts.AttachmentRepository,
	storage contracts.ObjectStorage,
) *Handler {
	return &Handler{repo: repo, attachments: attachments, storage: storage}
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

	case events.EventTypeUserJoinedRoom, events.EventTypeUserLeftRoom:
		// Postgres já foi atualizado pela API; fanout sincroniza Redis.

	default:
		log.Logger.Warn("unknown event type",
			"event_type", event.EventType,
			"event_id", event.EventID,
		)
	}
}

func (h *Handler) handleMessageSent(event events.Event) {
	var payload events.MessageSent

	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		log.Logger.Warn("persistence poison payload", "event_type", event.EventType, "event_id", event.EventID, "error", err)
		return
	}

	// regra mínima
	if payload.IsZeroLogging {
		log.Logger.Info("skipping zero logging message")
		return
	}

	id, err := uuid.Parse(payload.MessageID)
	if err != nil {
		log.Logger.Warn("persistence poison payload", "event_type", event.EventType, "event_id", event.EventID, "error", err)
		return
	}

	roomID, err := uuid.Parse(payload.RoomID)
	if err != nil {
		log.Logger.Warn("persistence poison payload", "event_type", event.EventType, "event_id", event.EventID, "error", err)
		return
	}

	senderID, err := uuid.Parse(payload.SenderID)
	if err != nil {
		log.Logger.Warn("persistence poison payload", "event_type", event.EventType, "event_id", event.EventID, "error", err)
		return
	}

	expiresAt := payload.ExpiresAt

	if expiresAt == nil && payload.TTL > 0 {
		t := payload.SentAt.Add(time.Duration(payload.TTL) * time.Second)
		expiresAt = &t
	}

	caption := strings.TrimSpace(payload.Content)
	attachmentKey := strings.TrimSpace(payload.AttachmentObjectKey)
	if payload.MessageType == "IMAGE" {
		if attachmentKey == "" && strings.HasPrefix(caption, "messages/") {
			attachmentKey = caption
			caption = ""
		}
	}

	msg := &domain.Message{
		ID:               id,
		RoomID:           roomID,
		SenderID:         senderID,
		Content:          caption,
		MessageType:      payload.MessageType,
		TTL:              payload.TTL,
		CreatedAt:        payload.SentAt,
		ExpiresAt:        expiresAt,
		DestroyAfterRead: payload.DestroyAfterRead,
	}

	if _, err := h.repo.Save(context.Background(), msg); err != nil {
		log.Logger.Error("failed to persist message", "error", err, "message_id", msg.ID)
		return
	}

	if payload.MessageType == "IMAGE" && attachmentKey != "" && h.attachments != nil {
		attachment := &domain.Attachment{
			ID:        uuid.New(),
			MessageID: id,
			RoomID:    roomID,
			ObjectKey: attachmentKey,
			Caption:   caption,
			MediaType: domain.MediaTypeImage,
			CreatedAt: payload.SentAt,
		}
		if err := h.attachments.ReplaceForMessage(context.Background(), attachment); err != nil {
			log.Logger.Error("failed to persist attachment", "error", err, "message_id", msg.ID)
		}
	}

	log.Logger.Info("message persisted", "message_id:", msg.ID)
}

func (h *Handler) handleMessageDeleted(event events.Event) {
	var payload events.MessageDeleted

	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		log.Logger.Warn("persistence poison payload", "event_type", event.EventType, "event_id", event.EventID, "error", err)
		return
	}

	id, err := uuid.Parse(payload.MessageID)
	if err != nil {
		log.Logger.Warn("persistence poison payload", "event_type", event.EventType, "event_id", event.EventID, "error", err)
		return
	}

	n, err := h.repo.Delete(context.Background(), id)
	if err != nil {
		log.Logger.Error("failed to delete message", "error", err)
		return
	}
	if n == 0 {
		log.Logger.Info("message delete noop already gone", "message_id", payload.MessageID)
	} else {
		log.Logger.Info("message deleted", "message_id", payload.MessageID)
	}

	h.purgeMedia(id)
}

func (h *Handler) handleMessageExpired(event events.Event) {
	var payload events.MessageExpired

	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		log.Logger.Warn("persistence poison payload", "event_type", event.EventType, "event_id", event.EventID, "error", err)
		return
	}

	id, err := uuid.Parse(payload.MessageID)
	if err != nil {
		log.Logger.Warn("persistence poison payload", "event_type", event.EventType, "event_id", event.EventID, "error", err)
		return
	}

	n, err := h.repo.Delete(context.Background(), id)
	if err != nil {
		log.Logger.Error("failed to delete expired message", "error", err)
		return
	}
	if n == 0 {
		log.Logger.Info("message expired noop already gone", "message_id", payload.MessageID)
	} else {
		log.Logger.Info("message expired removed", "message_id", payload.MessageID)
	}

	h.purgeMedia(id)
}

func (h *Handler) purgeMedia(messageID uuid.UUID) {
	if err := application.PurgeMessageMedia(context.Background(), h.attachments, h.storage, messageID); err != nil {
		log.Logger.Error("failed to purge message media", "message_id", messageID, "error", err)
	}
}
