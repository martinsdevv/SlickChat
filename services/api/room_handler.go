package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/martinsdevv/slickchat/core/auth"
	"github.com/martinsdevv/slickchat/core/contracts"
	"github.com/martinsdevv/slickchat/core/domain"
)

type RoomHandler struct {
	rooms           contracts.RoomRepository
	memberships     contracts.RoomMembershipRepository
	validateSession *auth.ValidateSessionUseCase
}

func NewRoomHandler(
	rooms contracts.RoomRepository,
	memberships contracts.RoomMembershipRepository,
	sessions contracts.SessionRepository,
) *RoomHandler {
	return &RoomHandler{
		rooms:           rooms,
		memberships:     memberships,
		validateSession: auth.NewValidateSessionUseCase(sessions),
	}
}

// POST /rooms
// Header: Authorization: Bearer <session_token>
// Body: {"type": "PUBLIC", "ttl": 0, "paranoid_mode": false, "zero_logging": false}
func (h *RoomHandler) CreateRoom(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	token := extractBearerToken(r)
	if token == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	session, err := h.validateSession.Execute(r.Context(), token)
	if err != nil {
		if errors.Is(err, domain.ErrSessionExpired) || errors.Is(err, domain.ErrSessionNotFound) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	var body struct {
		Type         string `json:"type"`
		TTL          int    `json:"ttl"`
		ParanoidMode bool   `json:"paranoid_mode"`
		ZeroLogging  bool   `json:"zero_logging"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	roomType := domain.RoomType(body.Type)
	switch roomType {
	case domain.RoomTypePublic, domain.RoomTypePrivate, domain.RoomTypeDirect, domain.RoomTypeTemporary:
	default:
		http.Error(w, "invalid room type", http.StatusBadRequest)
		return
	}

	now := time.Now().UTC()
	room := &domain.Room{
		ID:           uuid.New(),
		Type:         roomType,
		OwnerID:      session.UserID,
		TTL:          body.TTL,
		ParanoidMode: body.ParanoidMode,
		ZeroLogging:  body.ZeroLogging,
		CreatedAt:    now,
	}

	if body.TTL > 0 {
		exp := now.Add(time.Duration(body.TTL) * time.Second)
		room.ExpiresAt = &exp
	}

	if err := h.rooms.Save(r.Context(), room); err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	// creator automatically becomes ADMIN
	if err := h.memberships.Add(r.Context(), room.ID, session.UserID, domain.RoleAdmin); err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createRoomResponse{
		RoomID:       room.ID.String(),
		Type:         string(room.Type),
		OwnerID:      room.OwnerID.String(),
		TTL:          room.TTL,
		ParanoidMode: room.ParanoidMode,
		ZeroLogging:  room.ZeroLogging,
		CreatedAt:    room.CreatedAt,
		ExpiresAt:    room.ExpiresAt,
	})
}

type createRoomResponse struct {
	RoomID       string     `json:"room_id"`
	Type         string     `json:"type"`
	OwnerID      string     `json:"owner_id"`
	TTL          int        `json:"ttl"`
	ParanoidMode bool       `json:"paranoid_mode"`
	ZeroLogging  bool       `json:"zero_logging"`
	CreatedAt    time.Time  `json:"created_at"`
	ExpiresAt    *time.Time `json:"expires_at,omitempty"`
}
