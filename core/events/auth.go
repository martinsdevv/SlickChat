package events

import "time"

const (
	EventTypeUserRegistered = "user.registered.v1"
	EventTypeUserLoggedIn   = "user.logged_in.v1"
	EventTypeUserLoggedOut  = "user.logged_out.v1"
)

type UserRegistered struct {
	UserID       string    `json:"user_id"`
	Handle       string    `json:"handle"`
	RegisteredAt time.Time `json:"registered_at"`
}

type UserLoggedIn struct {
	UserID    string    `json:"user_id"`
	SessionID string    `json:"session_id"`
	LoggedAt  time.Time `json:"logged_at"`
}

type UserLoggedOut struct {
	UserID    string    `json:"user_id"`
	SessionID string    `json:"session_id"`
	LoggedAt  time.Time `json:"logged_at"`
}
