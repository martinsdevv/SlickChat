package api

import (
	"context"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/martinsdevv/slickchat/core/auth"
	"github.com/martinsdevv/slickchat/core/domain"
)

type contextKey string

const sessionUserIDKey contextKey = "session_user_id"

// AuthMiddleware validates the Bearer session token and injects the user ID
// into the request context. Returns 401 if the token is missing or invalid.
func AuthMiddleware(validateSession *auth.ValidateSessionUseCase, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := extractBearerToken(r)
		if token == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		session, err := validateSession.Execute(r.Context(), token)
		if err != nil {
			if errors.Is(err, domain.ErrSessionExpired) || errors.Is(err, domain.ErrSessionNotFound) {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		ctx := context.WithValue(r.Context(), sessionUserIDKey, session.UserID)
		next(w, r.WithContext(ctx))
	}
}

// UserIDFromContext extracts the authenticated user ID from the request context.
// Returns uuid.Nil if the middleware was not applied.
func UserIDFromContext(ctx context.Context) uuid.UUID {
	id, _ := ctx.Value(sessionUserIDKey).(uuid.UUID)
	return id
}
