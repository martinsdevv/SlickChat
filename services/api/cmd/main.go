package main

import (
	"net/http"

	"github.com/martinsdevv/slickchat/infrastructure/config"
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

	handler := api.NewMessageHandler(repo)
	roomContextHandler := api.NewRoomContextHandler(roomRepo, membershipRepo)
	messageContextHandler := api.NewMessageContextHandler(repo)

	http.HandleFunc("/messages", handler.GetMessages)
	http.HandleFunc("/room-context", roomContextHandler.GetRoomContext)
	http.HandleFunc("/message-context", messageContextHandler.GetMessageContext)

	port := ":8081"
	log.Logger.Info("API running on port" + port)

	err = http.ListenAndServe(port, nil)
	if err != nil {
		panic(err)
	}
}
