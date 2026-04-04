package ws

func sendToConnection(connectionID string, event MessageSent) {
	mu.Lock()
	client, ok := clients[connectionID]
	mu.Unlock()

	if !ok {
		return
	}

	client.Write(map[string]interface{}{
		"type":    "message_received",
		"payload": event,
	})
}
