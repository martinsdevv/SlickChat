package ws

func sendAck(client *Client) {
	client.Write(map[string]interface{}{
		"type": "message_ack",
		"payload": map[string]string{
			"status": "received",
		},
	})
}
