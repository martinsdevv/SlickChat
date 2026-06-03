package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/martinsdevv/slickchat/core/contracts"
	redisinfra "github.com/martinsdevv/slickchat/infrastructure/redis"
	"github.com/redis/go-redis/v9"
)

type UnreadHandler struct {
	rdb         *redis.Client
	memberships contracts.RoomMembershipRepository
}

func NewUnreadHandler(rdb *redis.Client, memberships contracts.RoomMembershipRepository) *UnreadHandler {
	return &UnreadHandler{rdb: rdb, memberships: memberships}
}

type roomUnreadItem struct {
	RoomID string `json:"room_id"`
	Count  int64  `json:"count"`
}

// List GET /rooms/unread
func (h *UnreadHandler) List(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := UserIDFromContext(r.Context())
	if userID == uuid.Nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	counts, err := redisinfra.ListUnreadByUser(r.Context(), h.rdb, userID.String())
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	items := make([]roomUnreadItem, 0, len(counts))
	for roomID, count := range counts {
		roomUUID, err := uuid.Parse(roomID)
		if err != nil {
			continue
		}
		if _, err := h.memberships.Get(r.Context(), roomUUID, userID); err != nil {
			continue
		}
		items = append(items, roomUnreadItem{RoomID: roomID, Count: count})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}

type clearUnreadBody struct {
	RoomID string `json:"room_id"`
}

// Clear POST /rooms/unread/clear
func (h *UnreadHandler) Clear(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := UserIDFromContext(r.Context())
	if userID == uuid.Nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	roomIDRaw := strings.TrimSpace(r.URL.Query().Get("room_id"))
	if roomIDRaw == "" {
		var body clearUnreadBody
		if err := json.NewDecoder(r.Body).Decode(&body); err == nil {
			roomIDRaw = strings.TrimSpace(body.RoomID)
		}
	}
	if roomIDRaw == "" {
		http.Error(w, "room_id required", http.StatusBadRequest)
		return
	}

	roomID, err := uuid.Parse(roomIDRaw)
	if err != nil {
		http.Error(w, "invalid room_id", http.StatusBadRequest)
		return
	}

	if _, err := h.memberships.Get(r.Context(), roomID, userID); err != nil {
		http.Error(w, "not a room member", http.StatusForbidden)
		return
	}

	if err := redisinfra.ClearRoomUnread(r.Context(), h.rdb, userID.String(), roomID.String()); err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
