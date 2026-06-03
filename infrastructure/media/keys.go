package media

import (
	"fmt"

	"github.com/google/uuid"
)

func MessageAttachmentKey(roomID, messageID, attachmentID uuid.UUID, ext string) string {
	if ext != "" && ext[0] != '.' {
		ext = "." + ext
	}
	return fmt.Sprintf("messages/%s/%s/%s%s", roomID, messageID, attachmentID, ext)
}

func RoomAvatarKey(roomID, uploadID uuid.UUID, ext string) string {
	if ext != "" && ext[0] != '.' {
		ext = "." + ext
	}
	return fmt.Sprintf("rooms/%s/avatar/%s%s", roomID, uploadID, ext)
}

func UserAvatarKey(userID, uploadID uuid.UUID, ext string) string {
	if ext != "" && ext[0] != '.' {
		ext = "." + ext
	}
	return fmt.Sprintf("users/%s/avatar/%s%s", userID, uploadID, ext)
}

func RoomBannerKey(roomID, uploadID uuid.UUID, ext string) string {
	if ext != "" && ext[0] != '.' {
		ext = "." + ext
	}
	return fmt.Sprintf("rooms/%s/banner/%s%s", roomID, uploadID, ext)
}
