package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/martinsdevv/slickchat/core/contracts"
	"github.com/martinsdevv/slickchat/core/domain"
)

const pgUniqueViolation = "23505"

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) contracts.UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Save(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (id, username, discriminator, password_hash, recovery_key_hash, paranoid_mode, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := r.db.ExecContext(ctx, query,
		user.ID,
		user.Username,
		user.Discriminator,
		user.PasswordHash,
		user.RecoveryKeyHash,
		user.ParanoidMode,
		user.CreatedAt,
	)

	if err != nil {
		var pgErr *pq.Error
		if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation {
			return domain.ErrHandleTaken
		}
		return err
	}

	return nil
}

func (r *UserRepository) GetByHandle(ctx context.Context, username, discriminator string) (*domain.User, error) {
	query := `
		SELECT id, username, discriminator, password_hash, recovery_key_hash, paranoid_mode, created_at
		FROM users
		WHERE username = $1 AND discriminator = $2
	`

	var (
		id              uuid.UUID
		uname           string
		disc            string
		passwordHash    sql.NullString
		recoveryKeyHash sql.NullString
		paranoidMode    bool
		createdAt       sql.NullTime
	)

	err := r.db.QueryRowContext(ctx, query, username, discriminator).Scan(
		&id,
		&uname,
		&disc,
		&passwordHash,
		&recoveryKeyHash,
		&paranoidMode,
		&createdAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrUserNotFound
	}

	if err != nil {
		return nil, err
	}

	return &domain.User{
		ID:              id,
		Username:        uname,
		Discriminator:   disc,
		PasswordHash:    passwordHash.String,
		RecoveryKeyHash: recoveryKeyHash.String,
		ParanoidMode:    paranoidMode,
		CreatedAt:       createdAt.Time.UTC(),
	}, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	query := `
		SELECT id, username, discriminator, password_hash, recovery_key_hash, paranoid_mode, created_at
		FROM users
		WHERE id = $1
	`

	var (
		uname           string
		disc            string
		passwordHash    sql.NullString
		recoveryKeyHash sql.NullString
		paranoidMode    bool
		createdAt       sql.NullTime
	)

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&id,
		&uname,
		&disc,
		&passwordHash,
		&recoveryKeyHash,
		&paranoidMode,
		&createdAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrUserNotFound
	}

	if err != nil {
		return nil, err
	}

	return &domain.User{
		ID:              id,
		Username:        uname,
		Discriminator:   disc,
		PasswordHash:    passwordHash.String,
		RecoveryKeyHash: recoveryKeyHash.String,
		ParanoidMode:    paranoidMode,
		CreatedAt:       createdAt.Time.UTC(),
	}, nil
}

func (r *UserRepository) HandleExists(ctx context.Context, username, discriminator string) (bool, error) {
	query := `SELECT 1 FROM users WHERE username = $1 AND discriminator = $2 LIMIT 1`

	var dummy int
	err := r.db.QueryRowContext(ctx, query, username, discriminator).Scan(&dummy)

	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	return true, nil
}
