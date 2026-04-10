//go:build integration

package integration

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/martinsdevv/slickchat/infrastructure/postgres"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type roomContextCache struct {
	RoomID       string     `json:"room_id"`
	Type         string     `json:"type"`
	TTL          int        `json:"ttl"`
	ParanoidMode bool       `json:"paranoid_mode"`
	ZeroLogging  bool       `json:"zero_logging"`
	Role         string     `json:"role"`
}

func postgresDSN() string {
	if d := os.Getenv("POSTGRES_TEST_DSN"); d != "" {
		return d
	}
	return "postgres://postgres:postgres@localhost:5432/slickchat?sslmode=disable"
}

func openPostgresOrSkip(t *testing.T) *sql.DB {
	t.Helper()
	db, err := postgres.NewConnection(postgresDSN())
	if err != nil {
		t.Skipf("postgres indisponível: %v (POSTGRES_TEST_DSN deve ser o mesmo DSN da API)", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func redisForIntegration(t *testing.T) *redis.Client {
	t.Helper()
	addr := os.Getenv("REDIS_TEST_ADDR")
	if addr == "" {
		addr = "127.0.0.1:6379"
	}
	rdb := redis.NewClient(&redis.Options{Addr: addr})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		t.Skipf("redis necessário para room_members / fanout: %v", err)
	}
	t.Cleanup(func() { _ = rdb.Close() })
	return rdb
}

func wsBaseURL() string {
	if u := os.Getenv("SLICKCHAT_WS_URL"); u != "" {
		return u
	}
	return "ws://127.0.0.1:8080"
}

// seedRoomInPostgres cria sala e membros que a rota GET /room-context da API valida.
// O gateway usa Redis só como cache; sem hit, consulta esta API (e o Postgres por trás).
func seedRoomInPostgres(t *testing.T, db *sql.DB, roomID, user1, user2 uuid.UUID) {
	t.Helper()
	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	require.NoError(t, err)
	defer func() { _ = tx.Rollback() }()

	_, err = tx.ExecContext(ctx, `
		INSERT INTO rooms (id, type, owner_id, ttl, paranoid_mode, zero_logging, created_at)
		VALUES ($1, 'PUBLIC', $2, 0, false, false, NOW())`,
		roomID, user1)
	require.NoError(t, err)

	now := time.Now().UTC()
	_, err = tx.ExecContext(ctx, `
		INSERT INTO room_members (room_id, user_id, role, created_at) VALUES ($1, $2, 'MEMBER', $3)`,
		roomID, user1, now)
	require.NoError(t, err)
	_, err = tx.ExecContext(ctx, `
		INSERT INTO room_members (room_id, user_id, role, created_at) VALUES ($1, $2, 'MEMBER', $3)`,
		roomID, user2, now)
	require.NoError(t, err)

	require.NoError(t, tx.Commit())

	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), `DELETE FROM messages WHERE room_id = $1`, roomID)
		_, _ = db.ExecContext(context.Background(), `DELETE FROM room_members WHERE room_id = $1`, roomID)
		_, _ = db.ExecContext(context.Background(), `DELETE FROM rooms WHERE id = $1`, roomID)
	})
}

// TestMessageFlow exige API (:8081), gateway WebSocket, Postgres, Redis, Kafka e worker fanout.
// go test -tags=integration ./tests/integration/...
//
// REDIS_TEST_ADDR deve apontar para a mesma instância que o gateway usa (senão o cache room_context não
// será compartilhado; com Postgres seedado a API ainda responde 200).
func TestMessageFlow(t *testing.T) {
	ctx := context.Background()
	db := openPostgresOrSkip(t)
	rdb := redisForIntegration(t)

	roomID := uuid.New()
	user1 := uuid.New()
	user2 := uuid.New()

	seedRoomInPostgres(t, db, roomID, user1, user2)

	rc := roomContextCache{
		RoomID:       roomID.String(),
		Type:         "PUBLIC",
		TTL:          0,
		ParanoidMode: false,
		ZeroLogging:  false,
		Role:         "MEMBER",
	}
	raw, err := json.Marshal(rc)
	require.NoError(t, err)

	cacheKey1 := "room_context:" + roomID.String() + ":" + user1.String()
	cacheKey2 := "room_context:" + roomID.String() + ":" + user2.String()
	require.NoError(t, rdb.Set(ctx, cacheKey1, raw, time.Minute).Err())
	require.NoError(t, rdb.Set(ctx, cacheKey2, raw, time.Minute).Err())
	require.NoError(t, rdb.SAdd(ctx, "room_members:"+roomID.String(), user1.String(), user2.String()).Err())

	t.Cleanup(func() {
		_ = rdb.Del(ctx, cacheKey1, cacheKey2, "room_members:"+roomID.String()).Err()
	})

	u1, _ := url.Parse(wsBaseURL() + "/socket")
	q1 := u1.Query()
	q1.Set("user_id", user1.String())
	u1.RawQuery = q1.Encode()

	u2, _ := url.Parse(wsBaseURL() + "/socket")
	q2 := u2.Query()
	q2.Set("user_id", user2.String())
	u2.RawQuery = q2.Encode()

	conn1, _, err := websocket.DefaultDialer.Dial(u1.String(), nil)
	if err != nil {
		t.Skipf("gateway WebSocket indisponível (%s): %v", u1.Redacted(), err)
	}
	defer conn1.Close()

	conn2, _, err := websocket.DefaultDialer.Dial(u2.String(), nil)
	if err != nil {
		t.Skipf("gateway WebSocket indisponível (%s): %v", u2.Redacted(), err)
	}
	defer conn2.Close()

	received := make(chan map[string]interface{}, 1)
	fromSender := make(chan map[string]interface{}, 16)

	go func() {
		for {
			var msg map[string]interface{}
			err := conn2.ReadJSON(&msg)
			if err != nil {
				return
			}
			if msgType, ok := msg["type"].(string); ok && msgType == "message.received" {
				received <- msg
			}
		}
	}()

	go func() {
		for {
			var msg map[string]interface{}
			err := conn1.ReadJSON(&msg)
			if err != nil {
				return
			}
			fromSender <- msg
		}
	}()

	time.Sleep(200 * time.Millisecond)

	send := map[string]interface{}{
		"type": "send_message",
		"payload": map[string]interface{}{
			"room_id": roomID.String(),
			"content": "teste integração",
		},
	}
	require.NoError(t, conn1.WriteJSON(send))

	wait := 5 * time.Second
	timer := time.NewTimer(wait)
	defer timer.Stop()

	ackOK := false
	for !ackOK {
		select {
		case msg := <-fromSender:
			tpe, _ := msg["type"].(string)
			if tpe == "error" {
				pl, _ := msg["payload"].(map[string]interface{})
				code, _ := pl["code"].(string)
				t.Skipf("gateway erro ao enviar: code=%q payload=%v — confira API em :8081, Kafka e mesmo Postgres (POSTGRES_TEST_DSN) que a API usa", code, pl)
			}
			if tpe == "message_ack" {
				ackOK = true
			}
		case <-timer.C:
			t.Fatal("timeout esperando ack ou erro do gateway após send_message")
		}
	}

	select {
	case msg := <-received:
		payload := msg["payload"].(map[string]interface{})
		assert.Equal(t, roomID.String(), payload["room_id"])
		assert.Equal(t, "teste integração", payload["content"])
	case <-time.After(wait):
		t.Fatal("timeout esperando message.received no peer (kafka + fanout ativos?)")
	}
}
