package main

import (
	"context"
	"testing"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockRedis struct {
	mock.Mock
}

func (m *MockRedis) SIsMember(ctx context.Context, key string, member interface{}) *redis.BoolCmd {
	args := m.Called(ctx, key, member)
	return args.Get(0).(*redis.BoolCmd)
}

func (m *MockRedis) SMembers(ctx context.Context, key string) *redis.StringSliceCmd {
	args := m.Called(ctx, key)
	return args.Get(0).(*redis.StringSliceCmd)
}

func (m *MockRedis) Publish(ctx context.Context, channel string, message interface{}) *redis.IntCmd {
	args := m.Called(ctx, channel, message)
	return args.Get(0).(*redis.IntCmd)
}

func TestIsUserInRoom(t *testing.T) {
	ctx := context.Background()

	t.Run("User is in room", func(t *testing.T) {
		mockRdb := new(MockRedis)
		boolCmd := &redis.BoolCmd{}
		boolCmd.SetVal(true)
		mockRdb.On("SIsMember", ctx, "room_members:room123", "user456").Return(boolCmd)

		result := isUserInRoom(mockRdb, "user456", "room123")
		assert.True(t, result)
		mockRdb.AssertExpectations(t)
	})

	t.Run("User is not in room", func(t *testing.T) {
		mockRdb := new(MockRedis)
		boolCmd := &redis.BoolCmd{}
		boolCmd.SetVal(false)
		mockRdb.On("SIsMember", ctx, "room_members:room123", "user456").Return(boolCmd)

		result := isUserInRoom(mockRdb, "user456", "room123")
		assert.False(t, result)
		mockRdb.AssertExpectations(t)
	})

	t.Run("Redis error", func(t *testing.T) {
		mockRdb := new(MockRedis)
		boolCmd := &redis.BoolCmd{}
		boolCmd.SetErr(assert.AnError)
		mockRdb.On("SIsMember", ctx, "room_members:room123", "user456").Return(boolCmd)

		result := isUserInRoom(mockRdb, "user456", "room123")
		assert.False(t, result)
		mockRdb.AssertExpectations(t)
	})
}

func TestHandleFanout(t *testing.T) {
	t.Skip("Unit test for handleFanout requires extensive mocking; better suited for integration tests")
}
