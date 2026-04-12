package main

import (
	"net/http"

	coreauth "github.com/martinsdevv/slickchat/core/auth"
	"github.com/martinsdevv/slickchat/infrastructure/config"
	kafkainfra "github.com/martinsdevv/slickchat/infrastructure/kafka"
	"github.com/martinsdevv/slickchat/infrastructure/log"
	"github.com/martinsdevv/slickchat/infrastructure/postgres"
	redisinfra "github.com/martinsdevv/slickchat/infrastructure/redis"
	"github.com/martinsdevv/slickchat/services/api"
)

func main() {
	dsn := config.LoadDBConfig()

	db, err := postgres.NewConnection(dsn.PGURL())
	if err != nil {
		panic(err)
	}

	// Repositories
	repo := postgres.NewMessageRepository(db)
	roomRepo := postgres.NewRoomRepository(db)
	membershipRepo := postgres.NewRoomMembershipRepository(db)
	userRepo := postgres.NewUserRepository(db)
	sessionRepo := postgres.NewSessionRepository(db)

	producer := kafkainfra.NewProducer("localhost:9092")

	// Redis (needed for connection notifier)
	rdb := redisinfra.NewClient()
	connNotifier := redisinfra.NewConnectionNotifier(rdb)

	// Auth use cases
	registerUC := coreauth.NewRegisterUseCase(userRepo)
	loginUC := coreauth.NewLoginUseCase(userRepo, sessionRepo)
	logoutUC := coreauth.NewLogoutUseCase(sessionRepo)

	// Handlers
	handler := api.NewMessageHandler(repo)
	roomContextHandler := api.NewRoomContextHandler(roomRepo, membershipRepo)
	messageContextHandler := api.NewMessageContextHandler(repo)
	roomMembershipWrite := api.NewRoomMembershipWriteHandler(producer, roomRepo, membershipRepo)
	authHandler := api.NewAuthHandler(registerUC, loginUC, logoutUC, connNotifier)

	// Routes
	http.HandleFunc("/register", authHandler.Register)
	http.HandleFunc("/login", authHandler.Login)
	http.HandleFunc("/logout", authHandler.Logout)

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
