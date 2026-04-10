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

type RoomMembershipRepository struct {
	db *sql.DB
}

func NewRoomMembershipRepository(db *sql.DB) contracts.RoomMembershipRepository {
	return &RoomMembershipRepository{db: db}
}

func (r *RoomMembershipRepository) Get(ctx context.Context, roomID uuid.UUID, userID uuid.UUID) (*domain.RoomMembership, error) {
	query := `
		SELECT user_id, role
		FROM room_members
		WHERE room_id = $1 AND user_id = $2
	`

	var (
		rawUserID uuid.UUID
		rawRole   string
	)

	if err := r.db.QueryRowContext(ctx, query, roomID, userID).Scan(&rawUserID, &rawRole); err != nil {
		return nil, err
	}

	role := domain.Role(rawRole)
	switch role {
	case domain.RoleAdmin, domain.RoleModerator, domain.RoleMember:
		// ok
	default:
		return nil, errors.New("invalid role")
	}

	return &domain.RoomMembership{
		UserID: rawUserID,
		Role:   role,
	}, nil
}

func (r *RoomMembershipRepository) Add(ctx context.Context, roomID uuid.UUID, userID uuid.UUID, role domain.Role) error {
	switch role {
	case domain.RoleAdmin, domain.RoleModerator, domain.RoleMember:
	default:
		return errors.New("invalid role")
	}

	query := `
		INSERT INTO room_members (room_id, user_id, role, created_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (room_id, user_id) DO UPDATE SET role = EXCLUDED.role
	`
	_, err := r.db.ExecContext(ctx, query, roomID, userID, string(role), time.Now().UTC())
	return err
}

func (r *RoomMembershipRepository) Remove(ctx context.Context, roomID uuid.UUID, userID uuid.UUID) error {
	query := `DELETE FROM room_members WHERE room_id = $1 AND user_id = $2`
	_, err := r.db.ExecContext(ctx, query, roomID, userID)
	return err
}

