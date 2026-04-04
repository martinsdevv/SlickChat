package ws

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	kafkainfra "github.com/martinsdevv/slickchat/infrastructure/kafka"
)

type MockRedis struct {
	mock.Mock
}

func (m *MockRedis) SIsMember(ctx context.Context, key string, member interface{}) *redis.BoolCmd {
	args := m.Called(ctx, key, member)
	return args.Get(0).(*redis.BoolCmd)
}

func (m *MockRedis) SAdd(ctx context.Context, key string, members ...interface{}) *redis.IntCmd {
	args := m.Called(ctx, key, members)
	return args.Get(0).(*redis.IntCmd)
}

func (m *MockRedis) HSet(ctx context.Context, key string, values ...interface{}) *redis.IntCmd {
	args := m.Called(ctx, key, values)
	return args.Get(0).(*redis.IntCmd)
}

func (m *MockRedis) SRem(ctx context.Context, key string, members ...interface{}) *redis.IntCmd {
	args := m.Called(ctx, key, members)
	return args.Get(0).(*redis.IntCmd)
}

func (m *MockRedis) Del(ctx context.Context, keys ...string) *redis.IntCmd {
	args := m.Called(ctx, keys)
	return args.Get(0).(*redis.IntCmd)
}

func (m *MockRedis) Subscribe(ctx context.Context, channels ...string) *redis.PubSub {
	args := m.Called(ctx, channels)
	return args.Get(0).(*redis.PubSub)
}

type MockProducer struct {
	mock.Mock
}

func (m *MockProducer) Publish(ctx context.Context, event interface{}) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func TestIsUserInRoom(t *testing.T) {
	ctx := context.Background()

	t.Run("User is in room", func(t *testing.T) {
		mockRdb := new(MockRedis)
		boolCmd := &redis.BoolCmd{}
		boolCmd.SetVal(true)
		mockRdb.On("SIsMember", ctx, "room_members:room1", "user1").Return(boolCmd)

		result := isUserInRoom(mockRdb, "user1", "room1")
		assert.True(t, result)
		mockRdb.AssertExpectations(t)
	})

	t.Run("User is not in room", func(t *testing.T) {
		mockRdb := new(MockRedis)
		boolCmd := &redis.BoolCmd{}
		boolCmd.SetVal(false)
		mockRdb.On("SIsMember", ctx, "room_members:room1", "user1").Return(boolCmd)

		result := isUserInRoom(mockRdb, "user1", "room1")
		assert.False(t, result)
		mockRdb.AssertExpectations(t)
	})
}

func TestSendError(t *testing.T) {
	// Create a mock client
	mockConn := new(MockConn)
	client := &Client{
		Conn: mockConn,
	}

	// Expect WriteJSON to be called with error structure
	mockConn.On("WriteJSON", map[string]interface{}{
		"type": "error",
		"payload": map[string]string{
			"code": "not_in_room",
		},
	}).Return(nil)

	sendError(client, "not_in_room")
	mockConn.AssertExpectations(t)
}

type MockConn struct {
	mock.Mock
}

func (m *MockConn) ReadJSON(v interface{}) error {
	args := m.Called(v)
	return args.Error(0)
}

func (m *MockConn) WriteJSON(v interface{}) error {
	args := m.Called(v)
	return args.Error(0)
}

func (m *MockConn) Close() error {
	args := m.Called()
	return args.Error(0)
}

func TestHandleWS(t *testing.T) {
	// This is a complex integration test, better suited for integration tests
	t.Skip("WebSocket handler test requires full HTTP server; better suited for integration tests")
}
