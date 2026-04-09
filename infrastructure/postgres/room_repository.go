package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/martinsdevv/slickchat/core/contracts"
	"github.com/martinsdevv/slickchat/core/domain"
)

type RoomRepository struct {
	db *sql.DB
}

func NewRoomRepository(db *sql.DB) contracts.RoomRepository {
	return &RoomRepository{db: db}
}

func (r *RoomRepository) GetByID(ctx context.Context, roomID uuid.UUID) (*domain.Room, error) {
	query := `
		SELECT id, type, owner_id, ttl, paranoid_mode, zero_logging, created_at, expires_at
		FROM rooms
		WHERE id = $1
	`

	var (
		id           uuid.UUID
		roomType     string
		ownerID      sql.NullString
		ttl          int
		paranoidMode bool
		zeroLogging  bool
		createdAt    time.Time
		expiresAt    sql.NullTime
	)

	if err := r.db.QueryRowContext(ctx, query, roomID).Scan(
		&id,
		&roomType,
		&ownerID,
		&ttl,
		&paranoidMode,
		&zeroLogging,
		&createdAt,
		&expiresAt,
	); err != nil {
		return nil, err
	}

	var owner uuid.UUID
	if ownerID.Valid {
		parsed, err := uuid.Parse(ownerID.String)
		if err != nil {
			return nil, err
		}
		owner = parsed
	}

	var exp *time.Time
	if expiresAt.Valid {
		t := expiresAt.Time.UTC()
		exp = &t
	}

	rt := domain.RoomType(roomType)
	switch rt {
	case domain.RoomTypePublic, domain.RoomTypePrivate, domain.RoomTypeDirect, domain.RoomTypeTemporary:
		// ok
	default:
		return nil, errors.New("invalid room type")
	}

	return &domain.Room{
		ID:           id,
		Type:         rt,
		OwnerID:      owner,
		TTL:          ttl,
		ParanoidMode: paranoidMode,
		ZeroLogging:  zeroLogging,
		CreatedAt:    createdAt.UTC(),
		ExpiresAt:    exp,
	}, nil
}

