package domain

import "errors"

var (
	ErrRoomExpired      = errors.New("room expired")
	ErrPermissionDenied = errors.New("permission denied")

	ErrHandleTaken        = errors.New("username and discriminator already taken")
	ErrInvalidHandle      = errors.New("invalid handle format")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrSessionNotFound    = errors.New("session not found")
	ErrSessionExpired     = errors.New("session expired")
	ErrUserNotFound       = errors.New("user not found")
)
