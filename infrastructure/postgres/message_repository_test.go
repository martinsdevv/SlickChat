package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/martinsdevv/slickchat/core/domain"
)

func TestMessageRepository_Save(t *testing.T) {
	dsn := "postgres://postgres:postgres@localhost:5432/slickchat?sslmode=disable"

	db, err := NewConnection(dsn)
	if err != nil {
		t.Fatalf("failed to connect db: %v", err)
	}

	defer db.Close()

	repo := NewMessageRepository(db)

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
		db.Exec("DELETE FROM messages WHERE id = $1", msg.ID)
	})

	err = repo.Save(context.Background(), msg)
	if err != nil {
		t.Fatalf("failed to save message: %v", err)
	}

	row := db.QueryRow("SELECT content FROM messages WHERE id = $1", msg.ID)

	var content string
	err = row.Scan(&content)

	if err != nil {
		t.Fatalf("failed to query message: %v", err)
	}

	if content != "hello test" {
		t.Fatalf("unexpected content: %s", content)
	}
}
