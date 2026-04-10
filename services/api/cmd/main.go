package main

import (
	"net/http"

	"github.com/martinsdevv/slickchat/infrastructure/config"
	kafkainfra "github.com/martinsdevv/slickchat/infrastructure/kafka"
	"github.com/martinsdevv/slickchat/infrastructure/log"
	"github.com/martinsdevv/slickchat/infrastructure/postgres"
	"github.com/martinsdevv/slickchat/services/api"
)

func main() {
	dsn := config.LoadDBConfig()

	db, err := postgres.NewConnection(dsn.PGURL())
	if err != nil {
		panic(err)
	}

	repo := postgres.NewMessageRepository(db)
	roomRepo := postgres.NewRoomRepository(db)
	membershipRepo := postgres.NewRoomMembershipRepository(db)
	producer := kafkainfra.NewProducer("localhost:9092")

	handler := api.NewMessageHandler(repo)
	roomContextHandler := api.NewRoomContextHandler(roomRepo, membershipRepo)
	messageContextHandler := api.NewMessageContextHandler(repo)
	roomMembershipWrite := api.NewRoomMembershipWriteHandler(producer, roomRepo, membershipRepo)

	http.HandleFunc("/messages", handler.GetMessages)
	http.HandleFunc("/room-context", roomContextHandler.GetRoomContext)
	http.HandleFunc("/message-context", messageContextHandler.GetMessageContext)
	http.HandleFunc("/room-members", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			roomMembershipWrite.JoinRoom(w, r)
		case http.MethodDelete:
			roomMembershipWrite.LeaveRoom(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	port := ":8081"
	log.Logger.Info("API running on port" + port)

	err = http.ListenAndServe(port, nil)
	if err != nil {
		panic(err)
	}
}
