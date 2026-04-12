package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/martinsdevv/slickchat/core/auth"
	"github.com/martinsdevv/slickchat/core/contracts"
	"github.com/martinsdevv/slickchat/core/domain"
)

type AuthHandler struct {
	register  *auth.RegisterUseCase
	login     *auth.LoginUseCase
	logout    *auth.LogoutUseCase
	notifier  contracts.ConnectionNotifier
}

func NewAuthHandler(
	register *auth.RegisterUseCase,
	login *auth.LoginUseCase,
	logout *auth.LogoutUseCase,
	notifier contracts.ConnectionNotifier,
) *AuthHandler {
	return &AuthHandler{
		register: register,
		login:    login,
		logout:   logout,
		notifier: notifier,
	}
}

// POST /register
// Body: {"username": "alice", "password": "secret123"}
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	result, err := h.register.Execute(r.Context(), auth.RegisterInput{
		Username: body.Username,
		Password: body.Password,
	})

	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidHandle):
			http.Error(w, err.Error(), http.StatusBadRequest)
		case errors.Is(err, domain.ErrInvalidCredentials):
			http.Error(w, err.Error(), http.StatusBadRequest)
		default:
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(registerResponse{
		Handle:      result.User.Handle(),
		RecoveryKey: result.RecoveryKeyRaw,
		CreatedAt:   result.User.CreatedAt,
	})
}

// POST /login
// Body: {"handle": "alice#0042", "password": "secret123"}
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var body struct {
		Handle   string `json:"handle"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	ip := extractIP(r)

	result, err := h.login.Execute(r.Context(), auth.LoginInput{
		Handle:   body.Handle,
		Password: body.Password,
		IP:       ip,
	})

	if err != nil {
		if errors.Is(err, domain.ErrInvalidCredentials) {
			http.Error(w, "invalid credentials", http.StatusUnauthorized)
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(loginResponse{
		Token:     result.TokenRaw,
		UserID:    result.User.ID.String(),
		Handle:    result.User.Handle(),
		ExpiresAt: result.Session.ExpiresAt,
	})
}

// POST /logout
// Header: Authorization: Bearer <token>
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	token := extractBearerToken(r)
	if token == "" {
		http.Error(w, "missing authorization token", http.StatusUnauthorized)
		return
	}

	result, err := h.logout.Execute(r.Context(), auth.LogoutInput{TokenRaw: token})
	if err != nil {
		if errors.Is(err, domain.ErrSessionNotFound) {
			http.Error(w, "session not found", http.StatusUnauthorized)
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	// Best-effort: force-close all active WebSocket connections for this user.
	// Failure here does not roll back the logout — session is already gone.
	_ = h.notifier.ForceDisconnectUser(r.Context(), result.UserID)

	w.WriteHeader(http.StatusNoContent)
}

type registerResponse struct {
	Handle      string    `json:"handle"`
	RecoveryKey string    `json:"recovery_key"`
	CreatedAt   time.Time `json:"created_at"`
}

type loginResponse struct {
	Token     string    `json:"token"`
	UserID    string    `json:"user_id"`
	Handle    string    `json:"handle"`
	ExpiresAt time.Time `json:"expires_at"`
}

func extractBearerToken(r *http.Request) string {
	header := r.Header.Get("Authorization")
	if !strings.HasPrefix(header, "Bearer ") {
		return ""
	}
	return strings.TrimPrefix(header, "Bearer ")
}

func extractIP(r *http.Request) string {
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		return strings.SplitN(forwarded, ",", 2)[0]
	}
	if real := r.Header.Get("X-Real-IP"); real != "" {
		return real
	}

	addr := r.RemoteAddr
	if idx := strings.LastIndex(addr, ":"); idx != -1 {
		return addr[:idx]
	}
	return addr
}
