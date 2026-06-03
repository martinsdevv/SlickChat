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
		INSERT INTO rooms (id, name, description, type, owner_id, ttl, paranoid_mode, zero_logging, avatar_object_key, banner_object_key, created_at, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	ownerID := sql.NullString{}
	if room.OwnerID != uuid.Nil {
		ownerID = sql.NullString{String: room.OwnerID.String(), Valid: true}
	}

	var expiresAt sql.NullTime
	if room.ExpiresAt != nil {
		expiresAt = sql.NullTime{Time: *room.ExpiresAt, Valid: true}
	}

	var avatarKey, bannerKey sql.NullString
	if room.AvatarObjectKey != "" {
		avatarKey = sql.NullString{String: room.AvatarObjectKey, Valid: true}
	}
	if room.BannerObjectKey != "" {
		bannerKey = sql.NullString{String: room.BannerObjectKey, Valid: true}
	}

	_, err := r.db.ExecContext(ctx, query,
		room.ID,
		room.Name,
		room.Description,
		string(room.Type),
		ownerID,
		room.TTL,
		room.ParanoidMode,
		room.ZeroLogging,
		avatarKey,
		bannerKey,
		room.CreatedAt,
		expiresAt,
	)
	return err
}

func (r *RoomRepository) GetByID(ctx context.Context, roomID uuid.UUID) (*domain.Room, error) {
	query := `
		SELECT id, name, description, type, owner_id, ttl, paranoid_mode, zero_logging, avatar_object_key, banner_object_key, created_at, expires_at
		FROM rooms
		WHERE id = $1
	`
	return r.scanRoom(r.db.QueryRowContext(ctx, query, roomID))
}

func (r *RoomRepository) ListPublic(ctx context.Context, limit int) ([]*domain.Room, error) {
	query := `
		SELECT id, name, description, type, owner_id, ttl, paranoid_mode, zero_logging, avatar_object_key, banner_object_key, created_at, expires_at
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

func (r *RoomRepository) ListByUser(ctx context.Context, userID uuid.UUID, limit int) ([]*domain.Room, error) {
	query := `
		SELECT ro.id, ro.name, ro.description, ro.type, ro.owner_id, ro.ttl, ro.paranoid_mode, ro.zero_logging, ro.avatar_object_key, ro.banner_object_key, ro.created_at, ro.expires_at
		FROM rooms ro
		INNER JOIN room_members rm ON rm.room_id = ro.id
		WHERE rm.user_id = $1
		  AND (ro.expires_at IS NULL OR ro.expires_at > NOW())
		ORDER BY ro.created_at DESC
		LIMIT $2
	`

	rows, err := r.db.QueryContext(ctx, query, userID, limit)
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
		description  string
		roomType     string
		ownerID      sql.NullString
		ttl          int
		paranoidMode    bool
		zeroLogging     bool
		avatarObjectKey sql.NullString
		bannerObjectKey sql.NullString
		createdAt       time.Time
		expiresAt       sql.NullTime
	)

	if err := row.Scan(
		&id,
		&name,
		&description,
		&roomType,
		&ownerID,
		&ttl,
		&paranoidMode,
		&zeroLogging,
		&avatarObjectKey,
		&bannerObjectKey,
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

	room := &domain.Room{
		ID:           id,
		Name:         name,
		Description:  description,
		Type:         rt,
		OwnerID:      owner,
		TTL:          ttl,
		ParanoidMode: paranoidMode,
		ZeroLogging:  zeroLogging,
		CreatedAt:    createdAt.UTC(),
		ExpiresAt:    exp,
	}
	if avatarObjectKey.Valid {
		room.AvatarObjectKey = avatarObjectKey.String
	}
	if bannerObjectKey.Valid {
		room.BannerObjectKey = bannerObjectKey.String
	}
	return room, nil
}

func (r *RoomRepository) SetAvatarObjectKey(ctx context.Context, roomID uuid.UUID, objectKey string) (string, error) {
	return r.setRoomObjectKey(ctx, roomID, "avatar_object_key", objectKey)
}

func (r *RoomRepository) SetBannerObjectKey(ctx context.Context, roomID uuid.UUID, objectKey string) (string, error) {
	return r.setRoomObjectKey(ctx, roomID, "banner_object_key", objectKey)
}

func (r *RoomRepository) setRoomObjectKey(ctx context.Context, roomID uuid.UUID, column, objectKey string) (string, error) {
	var previous sql.NullString
	selectQuery := "SELECT " + column + " FROM rooms WHERE id = $1"
	if err := r.db.QueryRowContext(ctx, selectQuery, roomID).Scan(&previous); err != nil {
		return "", err
	}

	updateQuery := "UPDATE rooms SET " + column + " = $2 WHERE id = $1"
	if _, err := r.db.ExecContext(ctx, updateQuery, roomID, nullableText(objectKey)); err != nil {
		return "", err
	}

	if previous.Valid {
		return previous.String, nil
	}
	return "", nil
}

func nullableText(value string) sql.NullString {
	if value == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: value, Valid: true}
}
