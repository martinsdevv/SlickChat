package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/martinsdevv/slickchat/core/contracts"
	"github.com/martinsdevv/slickchat/core/domain"
	"golang.org/x/crypto/bcrypt"
)

const sessionTTL = 24 * time.Hour

// authSalt is a fixed application-level salt mixed into IP hashing.
// This prevents cross-application rainbow tables for IPs while keeping
// the operation deterministic per deployment.
const authSalt = "slickchat:authsalt:v1"

type LoginUseCase struct {
	users    contracts.UserRepository
	sessions contracts.SessionRepository
}

func NewLoginUseCase(users contracts.UserRepository, sessions contracts.SessionRepository) *LoginUseCase {
	return &LoginUseCase{users: users, sessions: sessions}
}

type LoginInput struct {
	Handle   string // username#xxxx
	Password string
	IP       string // optional; pass "" to skip ip_hash
}

type LoginResult struct {
	User     *domain.User
	Session  *domain.Session
	TokenRaw string
}

func (uc *LoginUseCase) Execute(ctx context.Context, input LoginInput) (*LoginResult, error) {
	username, discriminator, err := parseHandle(input.Handle)
	if err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	user, err := uc.users.GetByHandle(ctx, username, discriminator)
	if err != nil {
		// Always return the same error to prevent user enumeration.
		return nil, domain.ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	tokenRaw, tokenHash, err := generateSessionToken()
	if err != nil {
		return nil, fmt.Errorf("generate session token: %w", err)
	}

	ipHash := ""
	if input.IP != "" {
		ipHash = hashIP(input.IP)
	}

	now := time.Now().UTC()
	session := &domain.Session{
		ID:        uuid.New(),
		UserID:    user.ID,
		TokenHash: tokenHash,
		CreatedAt: now,
		ExpiresAt: now.Add(sessionTTL),
		IPHash:    ipHash,
	}

	if err := uc.sessions.Save(ctx, session); err != nil {
		return nil, fmt.Errorf("save session: %w", err)
	}

	return &LoginResult{
		User:     user,
		Session:  session,
		TokenRaw: tokenRaw,
	}, nil
}

func parseHandle(handle string) (username, discriminator string, err error) {
	parts := strings.SplitN(handle, "#", 2)
	if len(parts) != 2 || parts[0] == "" || len(parts[1]) != 4 {
		return "", "", domain.ErrInvalidHandle
	}

	return strings.ToLower(parts[0]), parts[1], nil
}

func generateSessionToken() (raw, hashed string, err error) {
	b := make([]byte, 32)
	if _, err = rand.Read(b); err != nil {
		return "", "", err
	}

	raw = hex.EncodeToString(b)
	sum := sha256.Sum256([]byte(raw))
	hashed = hex.EncodeToString(sum[:])
	return raw, hashed, nil
}

func hashIP(ip string) string {
	sum := sha256.Sum256([]byte(ip + ":" + authSalt))
	return hex.EncodeToString(sum[:])
}
