package ws

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/martinsdevv/slickchat/core/events"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockProducer struct {
	mock.Mock
}

func (m *MockProducer) Publish(ctx context.Context, event interface{}) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

type MockClient struct {
	mock.Mock
}

func (m *MockClient) Write(v interface{}) error {
	args := m.Called(v)
	return args.Error(0)
}

func TestHandleSendMessage(t *testing.T) {
	producer := new(MockProducer)
	payload := SendMessagePayload{
		RoomID:  "room123",
		Content: "Hello, world!",
	}
	userID := "user456"

	// Expect Publish to be called with an event that has correct fields
	producer.On("Publish", context.Background(), mock.MatchedBy(func(event interface{}) bool {
		e, ok := event.(events.Event)
		if !ok {
			return false
		}
		assert.Equal(t, events.EventTypeMessageSent, e.Type)
		assert.NotEmpty(t, e.ID)
		assert.WithinDuration(t, time.Now(), e.Timestamp, time.Second)

		var msg events.MessageSent
		err := json.Unmarshal(e.Payload, &msg)
		assert.NoError(t, err)
		assert.Equal(t, "room123", msg.RoomID)
		assert.Equal(t, "user456", msg.SenderID)
		assert.Equal(t, "Hello, world!", msg.Content)
		assert.NotEmpty(t, msg.MessageID)
		return true
	})).Return(nil)

	handleSendMessage(producer, payload, userID)
	producer.AssertExpectations(t)
}

func TestSendAck(t *testing.T) {
	client := new(MockClient)

	client.On("Write", map[string]interface{}{
		"type": "message_ack",
		"payload": map[string]string{
			"status": "received",
		},
	}).Return(nil)

	sendAck(client)
	client.AssertExpectations(t)
}
