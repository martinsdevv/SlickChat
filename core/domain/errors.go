package domain

import "errors"

var (
	ErrRoomExpired      = errors.New("room expired")
	ErrPermissionDenied = errors.New("permission denied")
)
