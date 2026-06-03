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
	storage         contracts.ObjectStorage
	validateSession *auth.ValidateSessionUseCase
}

func NewRoomHandler(
	rooms contracts.RoomRepository,
	memberships contracts.RoomMembershipRepository,
	sessions contracts.SessionRepository,
	storage contracts.ObjectStorage,
) *RoomHandler {
	return &RoomHandler{
		rooms:           rooms,
		memberships:     memberships,
		storage:         storage,
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
		Name         string `json:"name"`
		Description  string `json:"description"`
		Type         string `json:"type"`
		TTL          int    `json:"ttl"`
		ParanoidMode bool   `json:"paranoid_mode"`
		ZeroLogging  bool   `json:"zero_logging"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if body.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
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
		Name:         body.Name,
		Description:  body.Description,
		Type:         roomType,
		OwnerID:      session.UserID,
		TTL:          body.TTL,
		ParanoidMode: body.ParanoidMode,
		ZeroLogging:  body.ZeroLogging,
		CreatedAt:    now,
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
		Name:         room.Name,
		Description:  room.Description,
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
	Name         string     `json:"name"`
	Description  string     `json:"description"`
	Type         string     `json:"type"`
	OwnerID      string     `json:"owner_id"`
	TTL          int        `json:"ttl"`
	ParanoidMode bool       `json:"paranoid_mode"`
	ZeroLogging  bool       `json:"zero_logging"`
	CreatedAt    time.Time  `json:"created_at"`
	ExpiresAt    *time.Time `json:"expires_at,omitempty"`
}

// GET /rooms?limit=50
func (h *RoomHandler) ListRooms(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := UserIDFromContext(r.Context())
	if userID == uuid.Nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	limit := 50
	rooms, err := h.rooms.ListByUser(r.Context(), userID, limit)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	type roomItem struct {
		RoomID       string     `json:"room_id"`
		Name         string     `json:"name"`
		Description  string     `json:"description"`
		Type         string     `json:"type"`
		OwnerID      string     `json:"owner_id,omitempty"`
		TTL          int        `json:"ttl"`
		ParanoidMode bool       `json:"paranoid_mode"`
		ZeroLogging  bool       `json:"zero_logging"`
		AvatarObjectKey string `json:"avatar_object_key,omitempty"`
		BannerObjectKey string `json:"banner_object_key,omitempty"`
		CreatedAt    time.Time  `json:"created_at"`
		ExpiresAt    *time.Time `json:"expires_at,omitempty"`
	}

	items := make([]roomItem, 0, len(rooms))
	for _, room := range rooms {
		item := roomItem{
			RoomID:       room.ID.String(),
			Name:         room.Name,
			Description:  room.Description,
			Type:         string(room.Type),
			TTL:          room.TTL,
			ParanoidMode: room.ParanoidMode,
			ZeroLogging:  room.ZeroLogging,
			AvatarObjectKey: room.AvatarObjectKey,
			BannerObjectKey: room.BannerObjectKey,
			CreatedAt:    room.CreatedAt,
			ExpiresAt:    room.ExpiresAt,
		}
		if room.OwnerID != uuid.Nil {
			item.OwnerID = room.OwnerID.String()
		}
		items = append(items, item)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}

// GET /room-members?room_id=
func (h *RoomHandler) ListRoomMembers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
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

	userID := UserIDFromContext(r.Context())
	if userID == uuid.Nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	if _, err := h.memberships.Get(r.Context(), roomID, userID); err != nil {
		http.Error(w, "not a room member", http.StatusForbidden)
		return
	}

	members, err := h.memberships.ListByRoom(r.Context(), roomID)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	type memberItem struct {
		UserID          string `json:"user_id"`
		Handle          string `json:"handle"`
		Role            string `json:"role"`
		AvatarObjectKey string `json:"avatar_object_key,omitempty"`
	}

	items := make([]memberItem, 0, len(members))
	for _, m := range members {
		items = append(items, memberItem{
			UserID:          m.UserID.String(),
			Handle:          m.Handle,
			Role:            string(m.Role),
			AvatarObjectKey: m.AvatarObjectKey,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}
