package api

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/martinsdevv/slickchat/core/contracts"
	"github.com/martinsdevv/slickchat/core/domain"
)

type MessageHandler struct {
	repo contracts.MessageRepository
}

func NewMessageHandler(repo contracts.MessageRepository) *MessageHandler {
	return &MessageHandler{repo: repo}
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
