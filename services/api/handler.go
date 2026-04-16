package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/martinsdevv/slickchat/core/contracts"
	"github.com/martinsdevv/slickchat/core/domain"
)

type MessageHandler struct {
	repo        contracts.MessageRepository
	memberships contracts.RoomMembershipRepository
}

type RoomContextHandler struct {
	rooms       contracts.RoomRepository
	memberships contracts.RoomMembershipRepository
}

type MessageContextHandler struct {
	repo contracts.MessageRepository
}

func NewMessageHandler(repo contracts.MessageRepository, memberships contracts.RoomMembershipRepository) *MessageHandler {
	return &MessageHandler{
		repo:        repo,
		memberships: memberships,
	}
}

func NewRoomContextHandler(rooms contracts.RoomRepository, memberships contracts.RoomMembershipRepository) *RoomContextHandler {
	return &RoomContextHandler{
		rooms:       rooms,
		memberships: memberships,
	}
}

func NewMessageContextHandler(repo contracts.MessageRepository) *MessageContextHandler {
	return &MessageContextHandler{repo: repo}
}

func (h *MessageHandler) GetMessages(w http.ResponseWriter, r *http.Request) {
	rawRoomID := r.URL.Query().Get("room_id")

	if rawRoomID == "" {
		http.Error(w, "room_id required", http.StatusBadRequest)
		return
	}

	roomID, err := uuid.Parse(rawRoomID)
	if err != nil {
		http.Error(w, "invalid room_id", http.StatusBadRequest)
		return
	}

	userID := UserIDFromContext(r.Context())
	if userID == uuid.Nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	if _, err := h.memberships.Get(r.Context(), roomID, userID); err != nil {
		http.Error(w, "not a room member", http.StatusForbidden)
		return
	}

	messages, err := h.repo.ListByRoom(r.Context(), roomID, 0)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := make([]MessageResponse, 0, len(messages))

	for _, msg := range messages {
		response = append(response, toResponse(msg))
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func toResponse(msg *domain.Message) MessageResponse {
	return MessageResponse{
		ID:        msg.ID.String(),
		SenderID:  msg.SenderID.String(),
		Content:   msg.Content,
		Type:      msg.MessageType,
		CreatedAt: msg.CreatedAt,
		ExpiresAt: msg.ExpiresAt,
	}
}

type RoomContextResponse struct {
	RoomID       string     `json:"room_id"`
	Type         string     `json:"type"`
	OwnerID      string     `json:"owner_id,omitempty"`
	TTL          int        `json:"ttl"`
	ParanoidMode bool       `json:"paranoid_mode"`
	ZeroLogging  bool       `json:"zero_logging"`
	ExpiresAt    *time.Time `json:"expires_at,omitempty"`
	Role         string     `json:"role"`
}

func (h *RoomContextHandler) GetRoomContext(w http.ResponseWriter, r *http.Request) {
	rawRoomID := r.URL.Query().Get("room_id")
	rawUserID := r.URL.Query().Get("user_id")

	if rawRoomID == "" || rawUserID == "" {
		http.Error(w, "room_id and user_id required", http.StatusBadRequest)
		return
	}

	roomID, err := uuid.Parse(rawRoomID)
	if err != nil {
		http.Error(w, "invalid room_id", http.StatusBadRequest)
		return
	}

	userID, err := uuid.Parse(rawUserID)
	if err != nil {
		http.Error(w, "invalid user_id", http.StatusBadRequest)
		return
	}

	room, err := h.rooms.GetByID(r.Context(), roomID)
	if err != nil {
		http.Error(w, "room not found", http.StatusNotFound)
		return
	}

	membership, err := h.memberships.Get(r.Context(), roomID, userID)
	if err != nil {
		http.Error(w, "not a room member", http.StatusForbidden)
		return
	}

	resp := RoomContextResponse{
		RoomID:       room.ID.String(),
		Type:         string(room.Type),
		OwnerID:      room.OwnerID.String(),
		TTL:          room.TTL,
		ParanoidMode: room.ParanoidMode,
		ZeroLogging:  room.ZeroLogging,
		ExpiresAt:    room.ExpiresAt,
		Role:         string(membership.Role),
	}

	json.NewEncoder(w).Encode(resp)
}

type MessageContextResponse struct {
	MessageID        string     `json:"message_id"`
	RoomID           string     `json:"room_id"`
	SenderID         string     `json:"sender_id"`
	DestroyAfterRead bool       `json:"destroy_after_read"`
	ExpiresAt        *time.Time `json:"expires_at,omitempty"`
}

func (h *MessageContextHandler) GetMessageContext(w http.ResponseWriter, r *http.Request) {
	rawMessageID := r.URL.Query().Get("message_id")
	if rawMessageID == "" {
		http.Error(w, "message_id required", http.StatusBadRequest)
		return
	}

	messageID, err := uuid.Parse(rawMessageID)
	if err != nil {
		http.Error(w, "invalid message_id", http.StatusBadRequest)
		return
	}

	msg, err := h.repo.GetByID(r.Context(), messageID)
	if err != nil {
		http.Error(w, "message not found", http.StatusNotFound)
		return
	}

	resp := MessageContextResponse{
		MessageID:        msg.ID.String(),
		RoomID:           msg.RoomID.String(),
		SenderID:         msg.SenderID.String(),
		DestroyAfterRead: msg.DestroyAfterRead,
		ExpiresAt:        msg.ExpiresAt,
	}

	json.NewEncoder(w).Encode(resp)
}
