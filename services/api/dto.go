package api

import "time"

type MessageResponse struct {
	ID                  string     `json:"id"`
	SenderID            string     `json:"sender_id"`
	Content             string     `json:"content"`
	Caption             string     `json:"caption,omitempty"`
	Type                string     `json:"type"`
	AttachmentObjectKey string     `json:"attachment_object_key,omitempty"`
	CreatedAt           time.Time  `json:"created_at"`
	ExpiresAt           *time.Time `json:"expires_at,omitempty"`
}
