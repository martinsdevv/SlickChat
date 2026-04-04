package ws

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	kafkainfra "github.com/martinsdevv/slickchat/infrastructure/kafka"
	"github.com/redis/go-redis/v9"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type Client struct {
	Conn *websocket.Conn
	Mu   sync.Mutex
}

var (
	clients = map[string]*Client{}
	mu      sync.Mutex
)

type Message struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type SendMessagePayload struct {
	RoomID  string `json:"room_id"`
	Content string `json:"content"`
}

type OutMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

func HandleWS(rdb *redis.Client, producer *kafkainfra.Producer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("erro no upgrade:", err)
			return
		}

		connectionID := uuid.New().String()

		userID := r.URL.Query().Get("user_id")
		if userID == "" {
			userID = "anonymous"
		}
		gatewayID := "gateway-1"

		client := &Client{
			Conn: conn,
		}

		mu.Lock()
		clients[connectionID] = client
		mu.Unlock()

		go subscribeConnection(rdb, connectionID)

		rdb.SAdd(r.Context(), "user_connections:"+userID, connectionID)

		rdb.HSet(r.Context(), "connection:"+connectionID, map[string]interface{}{
			"user_id":    userID,
			"gateway_id": gatewayID,
		})

		log.Println("Nova conexão:", connectionID)

		defer func() {

			mu.Lock()
			delete(clients, connectionID)
			mu.Unlock()

			rdb.SRem(r.Context(), "user_connections:"+userID, connectionID)
			rdb.Del(r.Context(), "connection:"+connectionID)

			conn.Close()
		}()

		for {
			var msg Message
			err := conn.ReadJSON(&msg)
			if err != nil {
				break
			}

			switch msg.Type {
			case "send_message":
				var payload SendMessagePayload
				json.Unmarshal(msg.Payload, &payload)

				if !isUserInRoom(rdb, userID, payload.RoomID) {
					sendError(client, "not_in_room")
					continue
				}

				handleSendMessage(producer, payload)
				sendAck(client)
			}
		}
	}
}

func subscribeConnection(rdb *redis.Client, connectionID string) {
	ctx := context.Background()

	pubsub := rdb.Subscribe(ctx, "connection:"+connectionID)

	ch := pubsub.Channel()

	for msg := range ch {
		var event MessageSent
		json.Unmarshal([]byte(msg.Payload), &event)

		sendToConnection(connectionID, event)
	}
}

func (c *Client) Write(v interface{}) error {
	c.Mu.Lock()
	defer c.Mu.Unlock()

	return c.Conn.WriteJSON(v)
}

func isUserInRoom(rdb *redis.Client, userID, roomID string) bool {
	ctx := context.Background()

	exists, err := rdb.SIsMember(ctx, "room_members:"+roomID, userID).Result()
	if err != nil {
		return false
	}

	return exists
}

func sendError(client *Client, code string) {
	client.Write(map[string]interface{}{
		"type": "error",
		"payload": map[string]string{
			"code": code,
		},
	})
}
