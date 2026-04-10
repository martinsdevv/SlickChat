package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRoom_CanDeleteMessage(t *testing.T) {
	now := time.Date(2026, 4, 1, 12, 0, 0, 0, time.UTC)
	expiredAt := now.Add(-time.Hour)
	owner := uuid.New()
	sender := uuid.New()
	other := uuid.New()
	roomID := uuid.New()

	t.Run("sala expirada nega", func(t *testing.T) {
		r := &Room{ID: roomID, OwnerID: owner}
		r.ExpiresAt = &expiredAt
		err := r.CanDeleteMessage(owner, &RoomMembership{UserID: owner, Role: RoleAdmin}, sender, now)
		assert.ErrorIs(t, err, ErrRoomExpired)
	})

	t.Run("dono da sala pode", func(t *testing.T) {
		r := &Room{ID: roomID, OwnerID: owner}
		err := r.CanDeleteMessage(owner, &RoomMembership{UserID: owner, Role: RoleMember}, sender, now)
		require.NoError(t, err)
	})

	t.Run("autor da mensagem pode", func(t *testing.T) {
		r := &Room{ID: roomID, OwnerID: owner}
		err := r.CanDeleteMessage(sender, &RoomMembership{UserID: sender, Role: RoleMember}, sender, now)
		require.NoError(t, err)
	})

	t.Run("moderador pode", func(t *testing.T) {
		r := &Room{ID: roomID, OwnerID: owner}
		mod := uuid.New()
		err := r.CanDeleteMessage(mod, &RoomMembership{UserID: mod, Role: RoleModerator}, sender, now)
		require.NoError(t, err)
	})

	t.Run("membro comum não pode apagar mensagem alheia", func(t *testing.T) {
		r := &Room{ID: roomID, OwnerID: owner}
		err := r.CanDeleteMessage(other, &RoomMembership{UserID: other, Role: RoleMember}, sender, now)
		assert.ErrorIs(t, err, ErrPermissionDenied)
	})

	t.Run("membership nil e não dono/autor nega", func(t *testing.T) {
		r := &Room{ID: roomID, OwnerID: owner}
		err := r.CanDeleteMessage(other, nil, sender, now)
		assert.ErrorIs(t, err, ErrPermissionDenied)
	})
}
