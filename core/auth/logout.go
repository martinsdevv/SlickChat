package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/google/uuid"
	"github.com/martinsdevv/slickchat/core/contracts"
	"github.com/martinsdevv/slickchat/core/domain"
)

type LogoutUseCase struct {
	sessions contracts.SessionRepository
}

func NewLogoutUseCase(sessions contracts.SessionRepository) *LogoutUseCase {
	return &LogoutUseCase{sessions: sessions}
}

type LogoutInput struct {
	TokenRaw string
}

type LogoutResult struct {
	UserID uuid.UUID
}

func (uc *LogoutUseCase) Execute(ctx context.Context, input LogoutInput) (*LogoutResult, error) {
	if input.TokenRaw == "" {
		return nil, domain.ErrSessionNotFound
	}

	sum := sha256.Sum256([]byte(input.TokenRaw))
	tokenHash := hex.EncodeToString(sum[:])

	session, err := uc.sessions.GetByTokenHash(ctx, tokenHash)
	if err != nil {
		return nil, domain.ErrSessionNotFound
	}

	if err := uc.sessions.DeleteByTokenHash(ctx, tokenHash); err != nil {
		return nil, fmt.Errorf("delete session: %w", err)
	}

	return &LogoutResult{UserID: session.UserID}, nil
}
