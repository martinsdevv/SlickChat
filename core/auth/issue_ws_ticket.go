package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/martinsdevv/slickchat/core/contracts"
)

const wsTicketTTL = 30 * time.Second

type IssueWSTicketUseCase struct {
	validateSession *ValidateSessionUseCase
	tickets         contracts.WSTicketStore
}

func NewIssueWSTicketUseCase(
	sessions contracts.SessionRepository,
	tickets contracts.WSTicketStore,
) *IssueWSTicketUseCase {
	return &IssueWSTicketUseCase{
		validateSession: NewValidateSessionUseCase(sessions),
		tickets:         tickets,
	}
}

type IssueWSTicketResult struct {
	TicketRaw string
}

func (uc *IssueWSTicketUseCase) Execute(ctx context.Context, sessionTokenRaw string) (*IssueWSTicketResult, error) {
	session, err := uc.validateSession.Execute(ctx, sessionTokenRaw)
	if err != nil {
		return nil, err
	}

	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return nil, fmt.Errorf("generate ws ticket: %w", err)
	}

	ticketRaw := hex.EncodeToString(b)
	sum := sha256.Sum256([]byte(ticketRaw))
	ticketHash := hex.EncodeToString(sum[:])

	if err := uc.tickets.Save(ctx, ticketHash, session.UserID, wsTicketTTL); err != nil {
		return nil, fmt.Errorf("save ws ticket: %w", err)
	}

	return &IssueWSTicketResult{TicketRaw: ticketRaw}, nil
}
