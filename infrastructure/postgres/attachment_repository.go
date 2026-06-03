package postgres

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/martinsdevv/slickchat/core/contracts"
	"github.com/martinsdevv/slickchat/core/domain"
)

type AttachmentRepository struct {
	db *sql.DB
}

func NewAttachmentRepository(db *sql.DB) contracts.AttachmentRepository {
	return &AttachmentRepository{db: db}
}

func (r *AttachmentRepository) Save(ctx context.Context, attachment *domain.Attachment) error {
	var caption sql.NullString
	if attachment.Caption != "" {
		caption = sql.NullString{String: attachment.Caption, Valid: true}
	}

	query := `
		INSERT INTO attachments (id, message_id, room_id, object_key, caption, media_type, size_bytes, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.db.ExecContext(ctx, query,
		attachment.ID,
		attachment.MessageID,
		attachment.RoomID,
		attachment.ObjectKey,
		caption,
		string(attachment.MediaType),
		attachment.SizeBytes,
		attachment.CreatedAt,
	)
	return err
}

func (r *AttachmentRepository) ReplaceForMessage(ctx context.Context, attachment *domain.Attachment) error {
	if _, err := r.DeleteByMessageID(ctx, attachment.MessageID); err != nil {
		return err
	}
	return r.Save(ctx, attachment)
}

func (r *AttachmentRepository) ListByMessageID(ctx context.Context, messageID uuid.UUID) ([]*domain.Attachment, error) {
	query := `
		SELECT id, message_id, room_id, object_key, caption, media_type, size_bytes, created_at
		FROM attachments
		WHERE message_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, messageID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*domain.Attachment
	for rows.Next() {
		var (
			id        uuid.UUID
			messageID uuid.UUID
			roomID    uuid.UUID
			objectKey string
			caption   sql.NullString
			mediaType string
			sizeBytes int64
			createdAt sql.NullTime
		)
		if err := rows.Scan(&id, &messageID, &roomID, &objectKey, &caption, &mediaType, &sizeBytes, &createdAt); err != nil {
			return nil, err
		}
		att := &domain.Attachment{
			ID:        id,
			MessageID: messageID,
			RoomID:    roomID,
			ObjectKey: objectKey,
			MediaType: domain.MediaType(mediaType),
			SizeBytes: sizeBytes,
			CreatedAt: createdAt.Time.UTC(),
		}
		if caption.Valid {
			att.Caption = caption.String
		}
		out = append(out, att)
	}
	return out, rows.Err()
}

func (r *AttachmentRepository) DeleteByMessageID(ctx context.Context, messageID uuid.UUID) ([]string, error) {
	query := `
		DELETE FROM attachments
		WHERE message_id = $1
		RETURNING object_key
	`
	rows, err := r.db.QueryContext(ctx, query, messageID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []string
	for rows.Next() {
		var key string
		if err := rows.Scan(&key); err != nil {
			return nil, err
		}
		keys = append(keys, key)
	}
	return keys, rows.Err()
}
