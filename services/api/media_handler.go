package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/martinsdevv/slickchat/core/contracts"
	"github.com/martinsdevv/slickchat/core/domain"
	"github.com/martinsdevv/slickchat/infrastructure/config"
	mediainfra "github.com/martinsdevv/slickchat/infrastructure/media"
)

type MediaHandler struct {
	rooms       contracts.RoomRepository
	memberships contracts.RoomMembershipRepository
	users       contracts.UserRepository
	storage     contracts.ObjectStorage
}

func NewMediaHandler(
	rooms contracts.RoomRepository,
	memberships contracts.RoomMembershipRepository,
	users contracts.UserRepository,
	storage contracts.ObjectStorage,
) *MediaHandler {
	return &MediaHandler{rooms: rooms, memberships: memberships, users: users, storage: storage}
}

type mediaUploadRequestBody struct {
	Purpose     string `json:"purpose"`
	RoomID      string `json:"room_id"`
	MessageID   string `json:"message_id,omitempty"`
	ContentType string `json:"content_type"`
	SizeBytes   int64  `json:"size_bytes"`
}

type mediaUploadRequestResponse struct {
	UploadID     string `json:"upload_id"`
	ObjectKey    string `json:"object_key"`
	UploadURL    string `json:"upload_url"`
	UploadViaAPI bool   `json:"upload_via_api"`
	ExpiresIn    int    `json:"expires_in_seconds"`
}

type mediaUploadCompleteBody struct {
	Purpose   string `json:"purpose"`
	RoomID    string `json:"room_id"`
	UploadID  string `json:"upload_id"`
	ObjectKey string `json:"object_key"`
}

type mediaUploadCompleteResponse struct {
	ObjectKey string `json:"object_key"`
	ReadURL   string `json:"read_url,omitempty"`
}

// UploadRequest POST /media/upload-request
func (h *MediaHandler) UploadRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := UserIDFromContext(r.Context())
	if userID == uuid.Nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var body mediaUploadRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := mediainfra.ValidateImageContentType(body.ContentType); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	purpose := strings.ToLower(strings.TrimSpace(body.Purpose))
	uploadID := uuid.New()
	ext := mediainfra.ExtensionForContentType(body.ContentType)

	var objectKey string
	switch purpose {
	case "user_avatar":
		if body.SizeBytes > mediainfra.MaxUserAvatarBytes {
			http.Error(w, "invalid file size for avatar", http.StatusBadRequest)
			return
		}
		objectKey = mediainfra.UserAvatarKey(userID, uploadID, ext)
	default:
		roomID, err := uuid.Parse(body.RoomID)
		if err != nil {
			http.Error(w, "invalid room_id", http.StatusBadRequest)
			return
		}

		_, err = h.rooms.GetByID(r.Context(), roomID)
		if err != nil {
			http.Error(w, "room not found", http.StatusNotFound)
			return
		}

		membership, err := h.memberships.Get(r.Context(), roomID, userID)
		if err != nil {
			http.Error(w, "not a room member", http.StatusForbidden)
			return
		}

		switch purpose {
		case "room_avatar":
			if membership.Role != domain.RoleAdmin {
				http.Error(w, "only admins can change room avatar", http.StatusForbidden)
				return
			}
			if body.SizeBytes > mediainfra.MaxRoomAvatarBytes {
				http.Error(w, "invalid file size for avatar", http.StatusBadRequest)
				return
			}
			objectKey = mediainfra.RoomAvatarKey(roomID, uploadID, ext)
		case "room_banner":
			if membership.Role != domain.RoleAdmin {
				http.Error(w, "only admins can change room banner", http.StatusForbidden)
				return
			}
			if body.SizeBytes > mediainfra.MaxRoomBannerBytes {
				http.Error(w, "invalid file size for banner", http.StatusBadRequest)
				return
			}
			objectKey = mediainfra.RoomBannerKey(roomID, uploadID, ext)
		case "message_image":
			if body.SizeBytes > mediainfra.MaxMessageImageBytes {
				http.Error(w, "invalid file size for image", http.StatusBadRequest)
				return
			}
			messageID, err := uuid.Parse(strings.TrimSpace(body.MessageID))
			if err != nil {
				http.Error(w, "invalid message_id", http.StatusBadRequest)
				return
			}
			attachmentID := uuid.New()
			objectKey = mediainfra.MessageAttachmentKey(roomID, messageID, attachmentID, ext)
		default:
			http.Error(w, "invalid purpose", http.StatusBadRequest)
			return
		}
	}

	mediaCfg := config.LoadMediaConfig()
	var uploadURL string
	expiresSec := mediaCfg.PresignUploadTTL
	if expiresSec <= 0 {
		expiresSec = 900
	}

	useProxy := mediainfra.ShouldUseProxyUpload(mediaCfg.PublicBaseURL)
	if useProxy {
		uploadURL = "/media/upload-put?object_key=" + url.QueryEscape(objectKey)
	} else {
		var expires time.Duration
		var err error
		uploadURL, expires, err = h.storage.PresignPut(r.Context(), objectKey, body.ContentType)
		if err != nil {
			http.Error(w, "media storage unavailable", http.StatusServiceUnavailable)
			return
		}
		uploadURL = mediainfra.PublicizePresignedURL(uploadURL, mediaCfg.PublicBaseURL)
		expiresSec = int(expires.Seconds())
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(mediaUploadRequestResponse{
		UploadID:     uploadID.String(),
		ObjectKey:    objectKey,
		UploadURL:    uploadURL,
		UploadViaAPI: useProxy,
		ExpiresIn:    expiresSec,
	})
}

// UploadPut PUT /media/upload-put?object_key=...
// Receives file bytes server-side (used when MinIO is not reachable from the browser, e.g. Cloudflare Tunnel).
func (h *MediaHandler) UploadPut(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := UserIDFromContext(r.Context())
	if userID == uuid.Nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	objectKey := strings.TrimSpace(r.URL.Query().Get("object_key"))
	if objectKey == "" {
		http.Error(w, "object_key required", http.StatusBadRequest)
		return
	}

	if err := h.authorizeUploadObjectKey(r.Context(), userID, objectKey); err != nil {
		if err == errUploadForbidden {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		http.Error(w, "invalid object_key", http.StatusBadRequest)
		return
	}

	contentType := strings.TrimSpace(r.Header.Get("Content-Type"))
	if err := mediainfra.ValidateImageContentType(contentType); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	maxBytes := mediainfra.MaxMessageImageBytes
	if mediainfra.IsUserAvatarKey(objectKey) {
		maxBytes = mediainfra.MaxUserAvatarBytes
	} else if strings.Contains(objectKey, "/banner/") {
		maxBytes = mediainfra.MaxRoomBannerBytes
	} else if strings.Contains(objectKey, "/avatar/") {
		maxBytes = mediainfra.MaxRoomAvatarBytes
	}

	limit := int64(maxBytes) + 1
	r.Body = http.MaxBytesReader(w, r.Body, limit)
	defer r.Body.Close()

	data, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "invalid upload body", http.StatusBadRequest)
		return
	}
	if int64(len(data)) > int64(maxBytes) {
		http.Error(w, "file too large", http.StatusBadRequest)
		return
	}
	if len(data) == 0 {
		http.Error(w, "empty body", http.StatusBadRequest)
		return
	}

	if err := h.storage.Put(r.Context(), objectKey, contentType, bytes.NewReader(data), int64(len(data))); err != nil {
		http.Error(w, "media storage unavailable", http.StatusServiceUnavailable)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// UploadComplete POST /media/upload-complete
func (h *MediaHandler) UploadComplete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := UserIDFromContext(r.Context())
	if userID == uuid.Nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var body mediaUploadCompleteBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if body.ObjectKey == "" {
		http.Error(w, "object_key required", http.StatusBadRequest)
		return
	}

	stat, err := h.storage.Stat(r.Context(), body.ObjectKey)
	if err != nil {
		http.Error(w, "uploaded object not found", http.StatusBadRequest)
		return
	}

	purpose := strings.ToLower(strings.TrimSpace(body.Purpose))
	maxBytes := mediainfra.MaxBytesForPurpose(purpose)
	if maxBytes > 0 {
		if err := mediainfra.ValidateStoredImageSize(stat.Size, maxBytes); err != nil {
			_ = h.storage.Delete(context.Background(), body.ObjectKey)
			http.Error(w, sizeErrorMessage(purpose), http.StatusBadRequest)
			return
		}
	}
	var previousKey string

	switch purpose {
	case "user_avatar":
		if !strings.HasPrefix(body.ObjectKey, "users/"+userID.String()+"/avatar/") {
			http.Error(w, "invalid object_key for user avatar", http.StatusBadRequest)
			return
		}
		previousKey, err = h.users.SetAvatarObjectKey(r.Context(), userID, body.ObjectKey)
		if err != nil {
			http.Error(w, "failed to update profile", http.StatusInternalServerError)
			return
		}
	case "message_image":
		roomID, err := uuid.Parse(body.RoomID)
		if err != nil {
			http.Error(w, "invalid room_id", http.StatusBadRequest)
			return
		}
		if !strings.HasPrefix(body.ObjectKey, "messages/"+roomID.String()+"/") {
			http.Error(w, "invalid object_key for message media", http.StatusBadRequest)
			return
		}
		if _, err := h.memberships.Get(r.Context(), roomID, userID); err != nil {
			http.Error(w, "not a room member", http.StatusForbidden)
			return
		}
	default:
		roomID, err := uuid.Parse(body.RoomID)
		if err != nil {
			http.Error(w, "invalid room_id", http.StatusBadRequest)
			return
		}
		if !strings.HasPrefix(body.ObjectKey, "rooms/"+roomID.String()+"/") {
			http.Error(w, "invalid object_key for room media", http.StatusBadRequest)
			return
		}

		membership, err := h.memberships.Get(r.Context(), roomID, userID)
		if err != nil {
			http.Error(w, "not a room member", http.StatusForbidden)
			return
		}

		switch purpose {
		case "room_avatar":
			if membership.Role != domain.RoleAdmin {
				http.Error(w, "only admins can change room avatar", http.StatusForbidden)
				return
			}
			previousKey, err = h.rooms.SetAvatarObjectKey(r.Context(), roomID, body.ObjectKey)
		case "room_banner":
			if membership.Role != domain.RoleAdmin {
				http.Error(w, "only admins can change room banner", http.StatusForbidden)
				return
			}
			previousKey, err = h.rooms.SetBannerObjectKey(r.Context(), roomID, body.ObjectKey)
		default:
			http.Error(w, "invalid purpose", http.StatusBadRequest)
			return
		}
		if err != nil {
			http.Error(w, "failed to update room", http.StatusInternalServerError)
			return
		}
	}

	if previousKey != "" && previousKey != body.ObjectKey {
		_ = h.storage.Delete(context.Background(), previousKey)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(mediaUploadCompleteResponse{ObjectKey: body.ObjectKey})
}

// ServeObject GET /media/object?object_key=
// Streams object from MinIO for authenticated room members (img tags cannot use presigned MinIO URLs reliably).
func (h *MediaHandler) ServeObject(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := UserIDFromContext(r.Context())
	if userID == uuid.Nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	objectKey := strings.TrimSpace(r.URL.Query().Get("object_key"))
	if objectKey == "" {
		http.Error(w, "invalid object_key", http.StatusBadRequest)
		return
	}

	switch {
	case mediainfra.IsUserAvatarKey(objectKey):
		// qualquer usuário autenticado pode ver avatares de perfil
	case mediainfra.IsRoomMediaKey(objectKey):
		roomID, ok := mediainfra.RoomIDFromObjectKey(objectKey)
		if !ok {
			http.Error(w, "invalid object_key", http.StatusBadRequest)
			return
		}
		if _, err := h.memberships.Get(r.Context(), roomID, userID); err != nil {
			http.Error(w, "not a room member", http.StatusForbidden)
			return
		}
	case mediainfra.IsMessageMediaKey(objectKey):
		roomID, ok := mediainfra.MessageRoomIDFromObjectKey(objectKey)
		if !ok {
			http.Error(w, "invalid object_key", http.StatusBadRequest)
			return
		}
		if _, err := h.memberships.Get(r.Context(), roomID, userID); err != nil {
			http.Error(w, "not a room member", http.StatusForbidden)
			return
		}
	default:
		http.Error(w, "invalid object_key", http.StatusBadRequest)
		return
	}

	reader, stat, err := h.storage.Get(r.Context(), objectKey)
	if err != nil {
		http.Error(w, "object not found", http.StatusNotFound)
		return
	}
	defer reader.Close()

	contentType := "application/octet-stream"
	if stat != nil && stat.ContentType != "" {
		contentType = stat.ContentType
	}
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", "private, max-age=300")
	_, _ = io.Copy(w, reader)
}

var errUploadForbidden = errors.New("upload forbidden")

func (h *MediaHandler) authorizeUploadObjectKey(ctx context.Context, userID uuid.UUID, objectKey string) error {
	switch {
	case mediainfra.IsUserAvatarKey(objectKey):
		ownerID, ok := mediainfra.UserIDFromObjectKey(objectKey)
		if !ok || ownerID != userID {
			return errUploadForbidden
		}
	case mediainfra.IsMessageMediaKey(objectKey):
		roomID, ok := mediainfra.MessageRoomIDFromObjectKey(objectKey)
		if !ok {
			return errors.New("invalid key")
		}
		if _, err := h.memberships.Get(ctx, roomID, userID); err != nil {
			return errUploadForbidden
		}
	case mediainfra.IsRoomMediaKey(objectKey):
		roomID, ok := mediainfra.RoomIDFromObjectKey(objectKey)
		if !ok {
			return errors.New("invalid key")
		}
		membership, err := h.memberships.Get(ctx, roomID, userID)
		if err != nil {
			return errUploadForbidden
		}
		if membership.Role != domain.RoleAdmin {
			return errUploadForbidden
		}
	default:
		return errors.New("invalid key")
	}
	return nil
}

func sizeErrorMessage(purpose string) string {
	switch purpose {
	case "room_banner":
		return "invalid file size for banner"
	case "message_image":
		return "invalid file size for image"
	default:
		return "invalid file size for avatar"
	}
}
