package persistence

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/martinsdevv/slickchat/core/domain"
	"github.com/martinsdevv/slickchat/core/events"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubMessageRepo struct {
	saveN   []int64
	saveI   int
	saveErr error
	delN    []int64
	delI    int
	delErr  error
	saved   []*domain.Message
}

func (s *stubMessageRepo) Save(ctx context.Context, msg *domain.Message) (int64, error) {
	if s.saveErr != nil {
		return 0, s.saveErr
	}
	var n int64
	if s.saveI < len(s.saveN) {
		n = s.saveN[s.saveI]
	}
	s.saveI++
	s.saved = append(s.saved, msg)
	return n, nil
}

func (s *stubMessageRepo) GetByID(ctx context.Context, messageID uuid.UUID) (*domain.Message, error) {
	return nil, nil
}

func (s *stubMessageRepo) ListByRoom(ctx context.Context, roomID uuid.UUID, limit int) ([]*domain.Message, error) {
	return nil, nil
}

func (s *stubMessageRepo) ListExpired(ctx context.Context, before time.Time, limit int) ([]*domain.Message, error) {
	return nil, nil
}

func (s *stubMessageRepo) Delete(ctx context.Context, messageID uuid.UUID) (int64, error) {
	if s.delErr != nil {
		return 0, s.delErr
	}
	var n int64
	if s.delI < len(s.delN) {
		n = s.delN[s.delI]
	}
	s.delI++
	return n, nil
}

func TestHandler_handleMessageSent_idempotenciaSaveZero(t *testing.T) {
	id := uuid.New()
	roomID := uuid.New()
	sender := uuid.New()
	repo := &stubMessageRepo{saveN: []int64{1, 0}}
	h := NewHandler(repo)

	payload := events.MessageSent{
		MessageID:     id.String(),
		RoomID:        roomID.String(),
		SenderID:      sender.String(),
		MessageType:   "TEXT",
		Content:       "x",
		SentAt:        time.Now().UTC(),
		IsZeroLogging: false,
	}
	ev1, err := events.NewEvent(events.EventTypeMessageSent, roomID.String(), payload)
	require.NoError(t, err)
	ev2, err := events.NewEvent(events.EventTypeMessageSent, roomID.String(), payload)
	require.NoError(t, err)

	h.Handle(ev1)
	h.Handle(ev2)

	assert.Len(t, repo.saved, 2)
	assert.Equal(t, 2, repo.saveI)
}

func TestHandler_handleMessageSent_zeroLoggingNaoPersiste(t *testing.T) {
	repo := &stubMessageRepo{saveN: []int64{1}}
	h := NewHandler(repo)
	id := uuid.New()
	roomID := uuid.New()
	sender := uuid.New()

	payload := events.MessageSent{
		MessageID:     id.String(),
		RoomID:        roomID.String(),
		SenderID:      sender.String(),
		MessageType:   "TEXT",
		Content:       "z",
		SentAt:        time.Now().UTC(),
		IsZeroLogging: true,
	}
	ev, err := events.NewEvent(events.EventTypeMessageSent, roomID.String(), payload)
	require.NoError(t, err)
	h.Handle(ev)
	assert.Empty(t, repo.saved)
	assert.Zero(t, repo.saveI)
}

func TestHandler_handleMessageDeleted_deleteIdempotenteZeroLinhas(t *testing.T) {
	id := uuid.New()
	roomID := uuid.New()
	repo := &stubMessageRepo{delN: []int64{0}}
	h := NewHandler(repo)

	payload := events.MessageDeleted{
		MessageID: id.String(),
		RoomID:    roomID.String(),
		DeletedAt: time.Now().UTC(),
	}
	ev, err := events.NewEvent(events.EventTypeMessageDeleted, roomID.String(), payload)
	require.NoError(t, err)
	h.Handle(ev)
	assert.Equal(t, 1, repo.delI)
}
