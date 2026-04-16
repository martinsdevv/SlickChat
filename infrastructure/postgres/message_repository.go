package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/martinsdevv/slickchat/core/contracts"
	"github.com/martinsdevv/slickchat/core/domain"
)

type MessageRepository struct {
	db *sql.DB
}

func NewMessageRepository(db *sql.DB) contracts.MessageRepository {
	return &MessageRepository{db: db}
}

func (r *MessageRepository) Save(ctx context.Context, msg *domain.Message) (int64, error) {
	query := `
		INSERT INTO messages (
			id,
			room_id,
			sender_id,
			content,
			message_type,
			ttl,
			destroy_after_read,
			created_at,
			expires_at
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		ON CONFLICT (id) DO NOTHING
	`

	res, err := r.db.ExecContext(
		ctx,
		query,
		msg.ID,
		msg.RoomID,
		msg.SenderID,
		msg.Content,
		msg.MessageType,
		msg.TTL,
		msg.DestroyAfterRead,
		msg.CreatedAt,
		msg.ExpiresAt,
	)
	if err != nil {
		return 0, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}
	return n, nil
}

func (r *MessageRepository) GetByID(ctx context.Context, messageID uuid.UUID) (*domain.Message, error) {
	query := `
		SELECT id, room_id, sender_id, content, message_type, ttl, destroy_after_read, created_at, expires_at
		FROM messages
		WHERE id = $1
	`

	var msg domain.Message
	if err := r.db.QueryRowContext(ctx, query, messageID).Scan(
		&msg.ID,
		&msg.RoomID,
		&msg.SenderID,
		&msg.Content,
		&msg.MessageType,
		&msg.TTL,
		&msg.DestroyAfterRead,
		&msg.CreatedAt,
		&msg.ExpiresAt,
	); err != nil {
		return nil, err
	}

	return &msg, nil
}

func (r *MessageRepository) ListByRoom(ctx context.Context, roomID uuid.UUID, limit int) ([]*domain.Message, error) {
	var (
		rows *sql.Rows
		err  error
	)
	if limit > 0 {
		query := `
			SELECT id, room_id, sender_id, content, message_type, ttl, destroy_after_read, created_at, expires_at
			FROM messages
			WHERE room_id = $1
			AND (expires_at IS NULL OR expires_at > NOW())
			ORDER BY created_at DESC
			LIMIT $2
		`
		rows, err = r.db.QueryContext(ctx, query, roomID, limit)
	} else {
		query := `
			SELECT id, room_id, sender_id, content, message_type, ttl, destroy_after_read, created_at, expires_at
			FROM messages
			WHERE room_id = $1
			AND (expires_at IS NULL OR expires_at > NOW())
			ORDER BY created_at DESC
		`
		rows, err = r.db.QueryContext(ctx, query, roomID)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*domain.Message

	for rows.Next() {
		var msg domain.Message

		err := rows.Scan(
			&msg.ID,
			&msg.RoomID,
			&msg.SenderID,
			&msg.Content,
			&msg.MessageType,
			&msg.TTL,
			&msg.DestroyAfterRead,
			&msg.CreatedAt,
			&msg.ExpiresAt,
		)

		if err != nil {
			return nil, err
		}

		messages = append(messages, &msg)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return messages, nil
}

func (r *MessageRepository) ListExpired(ctx context.Context, before time.Time, limit int) ([]*domain.Message, error) {
	query := `
		SELECT id, room_id, sender_id, content, message_type, ttl, destroy_after_read, created_at, expires_at
		FROM messages
		WHERE expires_at IS NOT NULL AND expires_at <= $1
		ORDER BY expires_at ASC
		LIMIT $2
	`

	rows, err := r.db.QueryContext(ctx, query, before, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*domain.Message

	for rows.Next() {
		var msg domain.Message

		err := rows.Scan(
			&msg.ID,
			&msg.RoomID,
			&msg.SenderID,
			&msg.Content,
			&msg.MessageType,
			&msg.TTL,
			&msg.DestroyAfterRead,
			&msg.CreatedAt,
			&msg.ExpiresAt,
		)

		if err != nil {
			return nil, err
		}

		messages = append(messages, &msg)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return messages, nil
}

func (r *MessageRepository) Delete(ctx context.Context, messageID uuid.UUID) (int64, error) {
	query := `DELETE FROM messages WHERE id = $1`

	res, err := r.db.ExecContext(ctx, query, messageID)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}
