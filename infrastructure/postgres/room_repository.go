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

func (r *RoomRepository) Save(ctx context.Context, room *domain.Room) error {
	query := `
		INSERT INTO rooms (id, name, type, owner_id, ttl, paranoid_mode, zero_logging, created_at, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	ownerID := sql.NullString{}
	if room.OwnerID != uuid.Nil {
		ownerID = sql.NullString{String: room.OwnerID.String(), Valid: true}
	}

	var expiresAt sql.NullTime
	if room.ExpiresAt != nil {
		expiresAt = sql.NullTime{Time: *room.ExpiresAt, Valid: true}
	}

	_, err := r.db.ExecContext(ctx, query,
		room.ID,
		room.Name,
		string(room.Type),
		ownerID,
		room.TTL,
		room.ParanoidMode,
		room.ZeroLogging,
		room.CreatedAt,
		expiresAt,
	)
	return err
}

func (r *RoomRepository) GetByID(ctx context.Context, roomID uuid.UUID) (*domain.Room, error) {
	query := `
		SELECT id, name, type, owner_id, ttl, paranoid_mode, zero_logging, created_at, expires_at
		FROM rooms
		WHERE id = $1
	`
	return r.scanRoom(r.db.QueryRowContext(ctx, query, roomID))
}

func (r *RoomRepository) ListPublic(ctx context.Context, limit int) ([]*domain.Room, error) {
	query := `
		SELECT id, name, type, owner_id, ttl, paranoid_mode, zero_logging, created_at, expires_at
		FROM rooms
		WHERE type = 'PUBLIC'
		  AND (expires_at IS NULL OR expires_at > NOW())
		ORDER BY created_at DESC
		LIMIT $1
	`

	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rooms []*domain.Room
	for rows.Next() {
		room, err := r.scanRoom(rows)
		if err != nil {
			return nil, err
		}
		rooms = append(rooms, room)
	}

	return rooms, rows.Err()
}

type roomScanner interface {
	Scan(dest ...any) error
}

func (r *RoomRepository) scanRoom(row roomScanner) (*domain.Room, error) {
	var (
		id           uuid.UUID
		name         string
		roomType     string
		ownerID      sql.NullString
		ttl          int
		paranoidMode bool
		zeroLogging  bool
		createdAt    time.Time
		expiresAt    sql.NullTime
	)

	if err := row.Scan(
		&id,
		&name,
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
	default:
		return nil, errors.New("invalid room type")
	}

	return &domain.Room{
		ID:           id,
		Name:         name,
		Type:         rt,
		OwnerID:      owner,
		TTL:          ttl,
		ParanoidMode: paranoidMode,
		ZeroLogging:  zeroLogging,
		CreatedAt:    createdAt.UTC(),
		ExpiresAt:    exp,
	}, nil
}
