package main

import (
	"context"
	"testing"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockRedisClient struct {
	mock.Mock
}

func (m *MockRedisClient) SIsMember(ctx context.Context, key string, member interface{}) *redis.BoolCmd {
	args := m.Called(ctx, key, member)
	return args.Get(0).(*redis.BoolCmd)
}

func (m *MockRedisClient) SMembers(ctx context.Context, key string) *redis.StringSliceCmd {
	args := m.Called(ctx, key)
	return args.Get(0).(*redis.StringSliceCmd)
}

func (m *MockRedisClient) Publish(ctx context.Context, channel string, message interface{}) *redis.IntCmd {
	args := m.Called(ctx, channel, message)
	return args.Get(0).(*redis.IntCmd)
}

func TestIsUserInRoom(t *testing.T) {
	ctx := context.Background()
	mockRdb := new(MockRedisClient)

	// Test case: user is in room
	boolCmd := &redis.BoolCmd{}
	boolCmd.SetVal(true)
	mockRdb.On("SIsMember", ctx, "room_members:room1", "user1").Return(boolCmd)

	result := isUserInRoom(mockRdb, "user1", "room1")
	assert.True(t, result)
	mockRdb.AssertExpectations(t)

	// Test case: user is not in room
	boolCmd2 := &redis.BoolCmd{}
	boolCmd2.SetVal(false)
	mockRdb2 := new(MockRedisClient)
	mockRdb2.On("SIsMember", ctx, "room_members:room1", "user1").Return(boolCmd2)

	result2 := isUserInRoom(mockRdb2, "user1", "room1")
	assert.False(t, result2)
	mockRdb2.AssertExpectations(t)
}

func TestHandleFanout(t *testing.T) {
	// This test is more complex due to multiple Redis calls
	// We'll create a more focused integration test in tests/integration
	// For unit test, we can skip or test with mocks
	t.Skip("Unit test for handleFanout requires extensive mocking; better suited for integration tests")
}
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
	cmd := args.Get(0).(*redis.BoolCmd)
	return cmd
}

func (m *MockRedis) SMembers(ctx context.Context, key string) *redis.StringSliceCmd {
	args := m.Called(ctx, key)
	cmd := args.Get(0).(*redis.StringSliceCmd)
	return cmd
}

func (m *MockRedis) Publish(ctx context.Context, channel string, message interface{}) *redis.IntCmd {
	args := m.Called(ctx, channel, message)
	cmd := args.Get(0).(*redis.IntCmd)
	return cmd
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
