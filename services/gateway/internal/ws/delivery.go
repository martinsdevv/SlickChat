package ws

import "github.com/martinsdevv/slickchat/core/events"

type WSMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

func sendToConnection(connectionID string, event events.MessageSent) {
	mu.Lock()
	client, ok := clients[connectionID]
	mu.Unlock()

	if !ok {
		return
	}

	client.Write(WSMessage{
		Type:    "message.received",
		Payload: event,
	})
}
