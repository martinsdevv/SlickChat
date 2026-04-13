package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

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
	users       contracts.UserRepository
}

func NewRoomMembershipWriteHandler(
	producer *kafkainfra.Producer,
	rooms contracts.RoomRepository,
	memberships contracts.RoomMembershipRepository,
	users contracts.UserRepository,
) *RoomMembershipWriteHandler {
	return &RoomMembershipWriteHandler{
		producer:    producer,
		rooms:       rooms,
		memberships: memberships,
		users:       users,
	}
}

// JoinRoom POST /room-members?room_id=
// Requires auth middleware — user_id comes from the session context.
func (h *RoomMembershipWriteHandler) JoinRoom(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := UserIDFromContext(r.Context())
	if userID == uuid.Nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	rawRoomID := r.URL.Query().Get("room_id")
	rawRole := r.URL.Query().Get("role")

	if rawRoomID == "" {
		http.Error(w, "room_id required", http.StatusBadRequest)
		return
	}

	roomID, err := uuid.Parse(rawRoomID)
	if err != nil {
		http.Error(w, "invalid room_id", http.StatusBadRequest)
		return
	}

	if rawRole != "" && domain.Role(rawRole) != domain.RoleMember {
		http.Error(w, "join only allows MEMBER role", http.StatusForbidden)
		return
	}

	err = application.JoinRoom(r.Context(), h.producer, h.rooms, h.memberships, roomID, userID, domain.RoleMember)
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

// AddMember POST /room-members/add
// Requires auth middleware. Caller must be ADMIN in room.
// Body: {"room_id":"...", "user_id":"..."} or {"room_id":"...", "handle":"user#1234"}
func (h *RoomMembershipWriteHandler) AddMember(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	requesterID := UserIDFromContext(r.Context())
	if requesterID == uuid.Nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var body struct {
		RoomID string `json:"room_id"`
		UserID string `json:"user_id"`
		Handle string `json:"handle"`
		Role   string `json:"role,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if body.RoomID == "" {
		http.Error(w, "room_id required", http.StatusBadRequest)
		return
	}

	roomID, err := uuid.Parse(body.RoomID)
	if err != nil {
		http.Error(w, "invalid room_id", http.StatusBadRequest)
		return
	}
	var targetUserID uuid.UUID
	if body.UserID != "" {
		targetUserID, err = uuid.Parse(body.UserID)
		if err != nil {
			http.Error(w, "invalid user_id", http.StatusBadRequest)
			return
		}
	} else if body.Handle != "" {
		username, discriminator, err := parseHandle(body.Handle)
		if err != nil {
			http.Error(w, "invalid handle", http.StatusBadRequest)
			return
		}
		user, err := h.users.GetByHandle(r.Context(), username, discriminator)
		if err != nil {
			if errors.Is(err, domain.ErrUserNotFound) {
				http.Error(w, "user not found", http.StatusNotFound)
				return
			}
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		targetUserID = user.ID
	} else {
		http.Error(w, "user_id or handle required", http.StatusBadRequest)
		return
	}

	requesterMembership, err := h.memberships.Get(r.Context(), roomID, requesterID)
	if err != nil {
		http.Error(w, "not a room member", http.StatusForbidden)
		return
	}
	if requesterMembership.Role != domain.RoleAdmin {
		http.Error(w, "only admins can add members", http.StatusForbidden)
		return
	}

	if _, err := h.memberships.Get(r.Context(), roomID, targetUserID); err == nil {
		http.Error(w, "user already in room", http.StatusConflict)
		return
	} else if !errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	role := domain.RoleMember
	if body.Role != "" {
		if domain.Role(body.Role) != domain.RoleMember {
			http.Error(w, "added users default to MEMBER role", http.StatusForbidden)
			return
		}
	}

	err = application.JoinRoom(r.Context(), h.producer, h.rooms, h.memberships, roomID, targetUserID, role)
	if err != nil {
		if errors.Is(err, application.ErrRoomNotFound) {
			http.Error(w, "room not found", http.StatusNotFound)
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func parseHandle(handle string) (username, discriminator string, err error) {
	parts := strings.SplitN(strings.TrimSpace(handle), "#", 2)
	if len(parts) != 2 || parts[0] == "" || len(parts[1]) != 4 {
		return "", "", domain.ErrInvalidHandle
	}
	return strings.ToLower(parts[0]), parts[1], nil
}

// LeaveRoom DELETE /room-members?room_id=
// Requires auth middleware — user_id comes from the session context.
func (h *RoomMembershipWriteHandler) LeaveRoom(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := UserIDFromContext(r.Context())
	if userID == uuid.Nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

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
