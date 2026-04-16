package api

import "time"

type MessageResponse struct {
	ID        string     `json:"id"`
	SenderID  string     `json:"sender_id"`
	Content   string     `json:"content"`
	Type      string     `json:"type"`
	CreatedAt time.Time  `json:"created_at"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}
