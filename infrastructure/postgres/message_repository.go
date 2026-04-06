package postgres

import (
	"context"
	"database/sql"

	"github.com/martinsdevv/slickchat/core/contracts"
	"github.com/martinsdevv/slickchat/core/domain"
)

type MessageRepository struct {
	db *sql.DB
}

func NewMessageRepository(db *sql.DB) contracts.MessageRepository {
	return &MessageRepository{db: db}
}

func (r *MessageRepository) Save(ctx context.Context, msg *domain.Message) error {
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
	`

	_, err := r.db.ExecContext(
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

	return err
}

func (r *MessageRepository) ListByRoom(ctx context.Context, roomID string, limit int) ([]*domain.Message, error) {
	query := `
		SELECT id, room_id, sender_id, content, message_type, ttl, destroy_after_read, created_at, expires_at
		FROM messages
		WHERE room_id = $1
		AND (expires_at IS NULL OR expires_at > NOW())
		ORDER BY created_at DESC
		LIMIT $2
	`

	rows, err := r.db.QueryContext(ctx, query, roomID, limit)
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

	return messages, nil
}
