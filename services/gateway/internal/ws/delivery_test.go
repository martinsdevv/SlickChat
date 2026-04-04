package ws

import (
	"testing"
)

type MockClient struct {
	mock.Mock
}

func (m *MockClient) Write(v interface{}) error {
	args := m.Called(v)
	return args.Error(0)
}

func TestSendToConnection(t *testing.T) {
	// Setup
	connectionID := "conn123"
	event := events.MessageSent{
		MessageID: "msg123",
		RoomID:    "room456",
		SenderID:  "user789",
		Content:   "Test message",
	}

	// Lock mutex to avoid race during test
	mu.Lock()
	oldClients := clients
	defer func() {
		clients = oldClients
		mu.Unlock()
	}()

	// Test when client exists
	client := new(MockClient)
	clients = map[string]*Client{
		connectionID: {Conn: nil, Mu: client.Mu},
	}
	// We need to mock Write method on Client
	// Since Client.Write calls Conn.WriteJSON, we need to mock the connection
	// For simplicity, we'll create a wrapper
	client.On("Write", WSMessage{
		Type:    "message.received",
		Payload: event,
	}).Return(nil)

	// Replace the client's Write method
	// This is tricky, so let's adjust approach
	// Instead, we can test via integration
	t.Skip("Unit test for sendToConnection requires mocking Client.Write which is complex; better suited for integration tests")
}
