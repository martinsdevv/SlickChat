package media

import (
	"fmt"
	"path"
	"strings"
)

const (
	MaxRoomAvatarBytes   = 5 << 20  // 5 MiB
	MaxUserAvatarBytes   = 5 << 20  // 5 MiB
	MaxRoomBannerBytes   = 8 << 20  // 8 MiB
	MaxMessageImageBytes = 15 << 20 // 15 MiB
)

var imageContentTypes = map[string]struct{}{
	"image/jpeg": {},
	"image/png":  {},
	"image/webp": {},
	"image/gif":  {},
}

func NormalizeImageContentType(contentType string) string {
	ct := strings.ToLower(strings.TrimSpace(contentType))
	if ct == "image/jpg" {
		return "image/jpeg"
	}
	return ct
}

func ValidateImageContentType(contentType string) error {
	if _, ok := imageContentTypes[NormalizeImageContentType(contentType)]; ok {
		return nil
	}
	return fmt.Errorf("unsupported content type: %s", contentType)
}

func ExtensionForContentType(contentType string) string {
	switch NormalizeImageContentType(contentType) {
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/webp":
		return ".webp"
	case "image/gif":
		return ".gif"
	default:
		return ""
	}
}

func MaxBytesForPurpose(purpose string) int64 {
	switch purpose {
	case "user_avatar":
		return MaxUserAvatarBytes
	case "room_avatar":
		return MaxRoomAvatarBytes
	case "room_banner":
		return MaxRoomBannerBytes
	case "message_image":
		return MaxMessageImageBytes
	default:
		return 0
	}
}

func ValidateStoredImageSize(size int64, maxBytes int64) error {
	if size <= 0 {
		return fmt.Errorf("empty file")
	}
	if size > maxBytes {
		return fmt.Errorf("file exceeds %d bytes", maxBytes)
	}
	return nil
}

func SanitizeFileName(name string) string {
	base := path.Base(strings.TrimSpace(name))
	if base == "." || base == "/" {
		return ""
	}
	return base
}
