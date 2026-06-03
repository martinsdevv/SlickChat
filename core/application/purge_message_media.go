package application

import (
	"context"

	"github.com/google/uuid"
	"github.com/martinsdevv/slickchat/core/contracts"
	"github.com/martinsdevv/slickchat/infrastructure/log"
)

// PurgeMessageMedia removes attachment rows and objects for a destroyed message.
func PurgeMessageMedia(
	ctx context.Context,
	attachments contracts.AttachmentRepository,
	storage contracts.ObjectStorage,
	messageID uuid.UUID,
) error {
	if attachments == nil || storage == nil {
		return nil
	}

	keys, err := attachments.DeleteByMessageID(ctx, messageID)
	if err != nil {
		return err
	}

	for _, key := range keys {
		if key == "" {
			continue
		}
		if err := storage.Delete(ctx, key); err != nil {
			log.Logger.Error("failed to delete attachment object", "object_key", key, "message_id", messageID, "error", err)
		}
	}

	return nil
}
