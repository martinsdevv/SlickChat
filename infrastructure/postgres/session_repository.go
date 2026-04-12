package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/martinsdevv/slickchat/core/contracts"
	"github.com/martinsdevv/slickchat/core/domain"
)

type SessionRepository struct {
	db *sql.DB
}

func NewSessionRepository(db *sql.DB) contracts.SessionRepository {
	return &SessionRepository{db: db}
}

func (r *SessionRepository) Save(ctx context.Context, session *domain.Session) error {
	query := `
		INSERT INTO sessions (id, user_id, token_hash, created_at, expires_at, ip_hash)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	ipHash := sql.NullString{String: session.IPHash, Valid: session.IPHash != ""}

	_, err := r.db.ExecContext(ctx, query,
		session.ID,
		session.UserID,
		session.TokenHash,
		session.CreatedAt,
		session.ExpiresAt,
		ipHash,
	)

	return err
}

func (r *SessionRepository) GetByTokenHash(ctx context.Context, tokenHash string) (*domain.Session, error) {
	query := `
		SELECT id, user_id, token_hash, created_at, expires_at, ip_hash
		FROM sessions
		WHERE token_hash = $1
	`

	var (
		id        uuid.UUID
		userID    uuid.UUID
		tHash     string
		createdAt sql.NullTime
		expiresAt sql.NullTime
		ipHash    sql.NullString
	)

	err := r.db.QueryRowContext(ctx, query, tokenHash).Scan(
		&id,
		&userID,
		&tHash,
		&createdAt,
		&expiresAt,
		&ipHash,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrSessionNotFound
	}

	if err != nil {
		return nil, err
	}

	return &domain.Session{
		ID:        id,
		UserID:    userID,
		TokenHash: tHash,
		CreatedAt: createdAt.Time.UTC(),
		ExpiresAt: expiresAt.Time.UTC(),
		IPHash:    ipHash.String,
	}, nil
}

func (r *SessionRepository) DeleteByTokenHash(ctx context.Context, tokenHash string) error {
	query := `DELETE FROM sessions WHERE token_hash = $1`

	result, err := r.db.ExecContext(ctx, query, tokenHash)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return domain.ErrSessionNotFound
	}

	return nil
}

func (r *SessionRepository) DeleteAllByUserID(ctx context.Context, userID uuid.UUID) error {
	query := `DELETE FROM sessions WHERE user_id = $1`

	_, err := r.db.ExecContext(ctx, query, userID)
	return err
}
