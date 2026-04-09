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
	repo contracts.MessageRepository
}

type RoomContextHandler struct {
	rooms       contracts.RoomRepository
	memberships contracts.RoomMembershipRepository
}

func NewMessageHandler(repo contracts.MessageRepository) *MessageHandler {
	return &MessageHandler{repo: repo}
}

func NewRoomContextHandler(rooms contracts.RoomRepository, memberships contracts.RoomMembershipRepository) *RoomContextHandler {
	return &RoomContextHandler{
		rooms:       rooms,
		memberships: memberships,
	}
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

	messages, err := h.repo.ListByRoom(r.Context(), roomID, 50)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var response []MessageResponse

	for _, msg := range messages {
		response = append(response, toResponse(msg))
	}

	json.NewEncoder(w).Encode(response)
}

func toResponse(msg *domain.Message) MessageResponse {
	return MessageResponse{
		ID:        msg.ID.String(),
		Content:   msg.Content,
		Type:      msg.MessageType,
		CreatedAt: msg.CreatedAt,
	}
}

type RoomContextResponse struct {
	RoomID       string     `json:"room_id"`
	Type         string     `json:"type"`
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
		TTL:          room.TTL,
		ParanoidMode: room.ParanoidMode,
		ZeroLogging:  room.ZeroLogging,
		ExpiresAt:    room.ExpiresAt,
		Role:         string(membership.Role),
	}

	json.NewEncoder(w).Encode(resp)
}
