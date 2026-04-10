package postgres

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/martinsdevv/slickchat/core/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func postgresDSN() string {
	if d := os.Getenv("POSTGRES_TEST_DSN"); d != "" {
		return d
	}
	return "postgres://postgres:postgres@localhost:5432/slickchat?sslmode=disable"
}

func openRepoOrSkip(t *testing.T) (repo *MessageRepository, db *sql.DB, cleanup func()) {
	t.Helper()
	var err error
	db, err = NewConnection(postgresDSN())
	if err != nil {
		t.Skipf("postgres indisponível: %v", err)
	}
	cleanup = func() { _ = db.Close() }
	repo = NewMessageRepository(db).(*MessageRepository)
	return repo, db, cleanup
}

func TestMessageRepository_Save(t *testing.T) {
	repo, db, cleanup := openRepoOrSkip(t)
	defer cleanup()

	msg := &domain.Message{
		ID:          uuid.New(),
		RoomID:      uuid.New(),
		SenderID:    uuid.New(),
		Content:     "hello test",
		MessageType: "TEXT",
		TTL:         0,
		CreatedAt:   time.Now(),
	}

	t.Cleanup(func() {
		_, _ = db.Exec("DELETE FROM messages WHERE id = $1", msg.ID)
	})

	n, err := repo.Save(context.Background(), msg)
	require.NoError(t, err)
	require.EqualValues(t, 1, n)

	var content string
	err = db.QueryRow("SELECT content FROM messages WHERE id = $1", msg.ID).Scan(&content)
	require.NoError(t, err)
	assert.Equal(t, "hello test", content)
}

func TestMessageRepository_Save_duplicateReturnsZero(t *testing.T) {
	repo, db, cleanup := openRepoOrSkip(t)
	defer cleanup()

	msg := &domain.Message{
		ID:          uuid.New(),
		RoomID:      uuid.New(),
		SenderID:    uuid.New(),
		Content:     "dup test",
		MessageType: "TEXT",
		TTL:         0,
		CreatedAt:   time.Now(),
	}

	t.Cleanup(func() {
		_, _ = db.Exec("DELETE FROM messages WHERE id = $1", msg.ID)
	})

	n1, err := repo.Save(context.Background(), msg)
	require.NoError(t, err)
	require.EqualValues(t, 1, n1)

	n2, err := repo.Save(context.Background(), msg)
	require.NoError(t, err)
	assert.Zero(t, n2, "ON CONFLICT DO NOTHING deve retornar 0 linhas")
}
