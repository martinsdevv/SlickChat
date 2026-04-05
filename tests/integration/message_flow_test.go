package integration

import (
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
)

func TestMessageFlow(t *testing.T) {

	u1 := "ws://localhost:8080/socket?user_id=1"
	conn1, _, err := websocket.DefaultDialer.Dial(u1, nil)
	assert.NoError(t, err)
	defer conn1.Close()

	u2 := "ws://localhost:8080/socket?user_id=2"
	conn2, _, err := websocket.DefaultDialer.Dial(u2, nil)
	assert.NoError(t, err)
	defer conn2.Close()

	received := make(chan map[string]interface{}, 1)

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

	time.Sleep(300 * time.Millisecond)

	send := map[string]interface{}{
		"type": "send_message",
		"payload": map[string]interface{}{
			"room_id": "1",
			"content": "teste integração",
		},
	}

	err = conn1.WriteJSON(send)
	assert.NoError(t, err)

	select {
	case msg := <-received:
		payload := msg["payload"].(map[string]interface{})

		assert.Equal(t, "1", payload["room_id"])
		assert.Equal(t, "teste integração", payload["content"])

	case <-time.After(5 * time.Second):
		t.Fatal("timeout esperando mensagem")
	}
}
