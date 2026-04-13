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
		SELECT rm.user_id,
		       CASE
		         WHEN ro.owner_id = rm.user_id THEN 'ADMIN'
		         ELSE rm.role
		       END AS effective_role
		FROM room_members rm
		JOIN rooms ro ON ro.id = rm.room_id
		WHERE rm.room_id = $1 AND rm.user_id = $2
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
		ON CONFLICT (room_id, user_id) DO NOTHING
	`
	_, err := r.db.ExecContext(ctx, query, roomID, userID, string(role), time.Now().UTC())
	return err
}

func (r *RoomMembershipRepository) Remove(ctx context.Context, roomID uuid.UUID, userID uuid.UUID) error {
	query := `DELETE FROM room_members WHERE room_id = $1 AND user_id = $2`
	_, err := r.db.ExecContext(ctx, query, roomID, userID)
	return err
}

func (r *RoomMembershipRepository) ListByRoom(ctx context.Context, roomID uuid.UUID) ([]contracts.MemberInfo, error) {
	query := `
		SELECT rm.user_id, u.username, u.discriminator,
		       CASE
		         WHEN ro.owner_id = rm.user_id THEN 'ADMIN'
		         ELSE rm.role
		       END AS effective_role
		FROM room_members rm
		JOIN users u ON u.id = rm.user_id
		JOIN rooms ro ON ro.id = rm.room_id
		WHERE rm.room_id = $1
		ORDER BY rm.created_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, roomID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []contracts.MemberInfo
	for rows.Next() {
		var (
			userID        uuid.UUID
			username      string
			discriminator string
			role          string
		)

		if err := rows.Scan(&userID, &username, &discriminator, &role); err != nil {
			return nil, err
		}

		members = append(members, contracts.MemberInfo{
			UserID: userID,
			Handle: username + "#" + discriminator,
			Role:   domain.Role(role),
		})
	}

	return members, rows.Err()
}

