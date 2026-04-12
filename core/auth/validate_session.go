package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"

	"github.com/martinsdevv/slickchat/core/contracts"
	"github.com/martinsdevv/slickchat/core/domain"
)

type ValidateSessionUseCase struct {
	sessions contracts.SessionRepository
}

func NewValidateSessionUseCase(sessions contracts.SessionRepository) *ValidateSessionUseCase {
	return &ValidateSessionUseCase{sessions: sessions}
}

// Execute validates a raw session token and returns the active session.
// Returns domain.ErrSessionNotFound or domain.ErrSessionExpired on failure.
func (uc *ValidateSessionUseCase) Execute(ctx context.Context, tokenRaw string) (*domain.Session, error) {
	if tokenRaw == "" {
		return nil, domain.ErrSessionNotFound
	}

	sum := sha256.Sum256([]byte(tokenRaw))
	tokenHash := hex.EncodeToString(sum[:])

	session, err := uc.sessions.GetByTokenHash(ctx, tokenHash)
	if err != nil {
		return nil, domain.ErrSessionNotFound
	}

	if session.IsExpired() {
		return nil, domain.ErrSessionExpired
	}

	return session, nil
}
