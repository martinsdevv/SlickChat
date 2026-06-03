package ws

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/martinsdevv/slickchat/core/contracts"
	"github.com/martinsdevv/slickchat/core/events"
	kafkainfra "github.com/martinsdevv/slickchat/infrastructure/kafka"
	"github.com/martinsdevv/slickchat/infrastructure/log"
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
	RoomID      string `json:"room_id"`
	Content     string `json:"content"`
	MessageID   string `json:"message_id,omitempty"`
	MessageType string `json:"message_type,omitempty"`
	ObjectKey   string `json:"object_key,omitempty"`
}

type OutMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

type MessageDeliveredPayload struct {
	MessageID string `json:"message_id"`
	RoomID    string `json:"room_id"`
}

type MessageReadPayload struct {
	MessageID string `json:"message_id"`
	RoomID    string `json:"room_id"`
}

type MessageDeletePayload struct {
	MessageID string `json:"message_id"`
	RoomID    string `json:"room_id"`
}

func HandleWS(rdb *redis.Client, producer *kafkainfra.Producer, tickets contracts.WSTicketStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ticketRaw := r.URL.Query().Get("ticket")
		if ticketRaw == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		sum := sha256.Sum256([]byte(ticketRaw))
		ticketHash := hex.EncodeToString(sum[:])

		resolvedUserID, err := tickets.GetAndDelete(r.Context(), ticketHash)
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Logger.Info("erro no upgrade", "error", err)
			return
		}

		connectionID := uuid.New().String()

		client := &Client{
			Conn: conn,
		}

		userID := resolvedUserID.String()
		gatewayID := "gateway-1"

		mu.Lock()
		clients[connectionID] = client
		mu.Unlock()

		go subscribeConnection(rdb, connectionID)

		rdb.SAdd(r.Context(), "user_connections:"+userID, connectionID)

		rdb.HSet(r.Context(), "connection:"+connectionID, map[string]interface{}{
			"user_id":    userID,
			"gateway_id": gatewayID,
		})

		log.Logger.Info("connection_opened",
			"connection_id", connectionID,
		)

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

				handleSendMessage(rdb, producer, client, userID, payload)
			case "message_delivered":
				var payload MessageDeliveredPayload
				json.Unmarshal(msg.Payload, &payload)

				handleMessageDelivered(rdb, producer, client, userID, payload)
			case "message_read":
				var payload MessageReadPayload
				json.Unmarshal(msg.Payload, &payload)

				handleMessageRead(rdb, producer, client, userID, payload)

			case "delete_message":
				var payload MessageDeletePayload
				err := json.Unmarshal(msg.Payload, &payload)
				if err != nil {
					log.Logger.Error("error in the payload", "error", err)
				}

				handleDeleteMessage(rdb, producer, client, userID, payload)
			}
		}
	}
}

func subscribeConnection(rdb *redis.Client, connectionID string) {
	ctx := context.Background()

	pubsub := rdb.Subscribe(ctx, "connection:"+connectionID)
	defer pubsub.Close()

	ch := pubsub.Channel()

	for msg := range ch {

		var event events.Event
		json.Unmarshal([]byte(msg.Payload), &event)

		handleIncomingEvent(connectionID, event)
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
