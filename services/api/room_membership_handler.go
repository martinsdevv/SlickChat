package api

import (
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/martinsdevv/slickchat/core/application"
	"github.com/martinsdevv/slickchat/core/contracts"
	"github.com/martinsdevv/slickchat/core/domain"
	kafkainfra "github.com/martinsdevv/slickchat/infrastructure/kafka"
)

type RoomMembershipWriteHandler struct {
	producer    *kafkainfra.Producer
	rooms       contracts.RoomRepository
	memberships contracts.RoomMembershipRepository
}

func NewRoomMembershipWriteHandler(
	producer *kafkainfra.Producer,
	rooms contracts.RoomRepository,
	memberships contracts.RoomMembershipRepository,
) *RoomMembershipWriteHandler {
	return &RoomMembershipWriteHandler{
		producer:    producer,
		rooms:       rooms,
		memberships: memberships,
	}
}

// JoinRoom POST /room-members?room_id=&user_id=&role=MEMBER
func (h *RoomMembershipWriteHandler) JoinRoom(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	rawRoomID := r.URL.Query().Get("room_id")
	rawUserID := r.URL.Query().Get("user_id")
	rawRole := r.URL.Query().Get("role")
	if rawRole == "" {
		rawRole = string(domain.RoleMember)
	}

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

	role := domain.Role(rawRole)
	switch role {
	case domain.RoleAdmin, domain.RoleModerator, domain.RoleMember:
	default:
		http.Error(w, "invalid role", http.StatusBadRequest)
		return
	}

	err = application.JoinRoom(r.Context(), h.producer, h.rooms, h.memberships, roomID, userID, role)
	if err != nil {
		if errors.Is(err, application.ErrRoomNotFound) {
			http.Error(w, "room not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// LeaveRoom DELETE /room-members?room_id=&user_id=
func (h *RoomMembershipWriteHandler) LeaveRoom(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

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

	err = application.LeaveRoom(r.Context(), h.producer, h.memberships, roomID, userID)
	if err != nil {
		if errors.Is(err, application.ErrNotRoomMember) {
			http.Error(w, "not a room member", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
