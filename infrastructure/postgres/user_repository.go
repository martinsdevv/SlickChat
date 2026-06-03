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
	var avatarKey sql.NullString
	if user.AvatarObjectKey != "" {
		avatarKey = sql.NullString{String: user.AvatarObjectKey, Valid: true}
	}

	query := `
		INSERT INTO users (id, username, discriminator, password_hash, recovery_key_hash, paranoid_mode, avatar_object_key, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.db.ExecContext(ctx, query,
		user.ID,
		user.Username,
		user.Discriminator,
		user.PasswordHash,
		user.RecoveryKeyHash,
		user.ParanoidMode,
		avatarKey,
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
	return r.scanUser(r.db.QueryRowContext(ctx, `
		SELECT id, username, discriminator, password_hash, recovery_key_hash, paranoid_mode, avatar_object_key, created_at
		FROM users
		WHERE username = $1 AND discriminator = $2
	`, username, discriminator))
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	user, err := r.scanUser(r.db.QueryRowContext(ctx, `
		SELECT id, username, discriminator, password_hash, recovery_key_hash, paranoid_mode, avatar_object_key, created_at
		FROM users
		WHERE id = $1
	`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrUserNotFound
	}
	return user, err
}

func (r *UserRepository) scanUser(row *sql.Row) (*domain.User, error) {
	var (
		id              uuid.UUID
		uname           string
		disc            string
		passwordHash    sql.NullString
		recoveryKeyHash sql.NullString
		paranoidMode    bool
		avatarObjectKey sql.NullString
		createdAt       sql.NullTime
	)

	err := row.Scan(
		&id,
		&uname,
		&disc,
		&passwordHash,
		&recoveryKeyHash,
		&paranoidMode,
		&avatarObjectKey,
		&createdAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrUserNotFound
	}

	if err != nil {
		return nil, err
	}

	user := &domain.User{
		ID:              id,
		Username:        uname,
		Discriminator:   disc,
		PasswordHash:    passwordHash.String,
		RecoveryKeyHash: recoveryKeyHash.String,
		ParanoidMode:    paranoidMode,
		CreatedAt:       createdAt.Time.UTC(),
	}
	if avatarObjectKey.Valid {
		user.AvatarObjectKey = avatarObjectKey.String
	}
	return user, nil
}

func (r *UserRepository) SetAvatarObjectKey(ctx context.Context, userID uuid.UUID, objectKey string) (string, error) {
	var previous sql.NullString
	if err := r.db.QueryRowContext(ctx, `
		SELECT avatar_object_key FROM users WHERE id = $1
	`, userID).Scan(&previous); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", domain.ErrUserNotFound
		}
		return "", err
	}

	var avatarKey sql.NullString
	if objectKey != "" {
		avatarKey = sql.NullString{String: objectKey, Valid: true}
	}
	if _, err := r.db.ExecContext(ctx, `
		UPDATE users SET avatar_object_key = $2 WHERE id = $1
	`, userID, avatarKey); err != nil {
		return "", err
	}
	if previous.Valid {
		return previous.String, nil
	}
	return "", nil
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
