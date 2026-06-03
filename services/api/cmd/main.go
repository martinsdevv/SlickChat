package main

import (
	"net/http"

	coreauth "github.com/martinsdevv/slickchat/core/auth"
	"github.com/martinsdevv/slickchat/infrastructure/config"
	kafkainfra "github.com/martinsdevv/slickchat/infrastructure/kafka"
	"github.com/martinsdevv/slickchat/infrastructure/log"
	"github.com/martinsdevv/slickchat/infrastructure/media"
	"github.com/martinsdevv/slickchat/infrastructure/postgres"
	redisinfra "github.com/martinsdevv/slickchat/infrastructure/redis"
	"github.com/martinsdevv/slickchat/services/api"
)

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

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

	producer := kafkainfra.NewProducer(config.KafkaBroker())

	// Redis
	rdb := redisinfra.NewClient()
	connNotifier := redisinfra.NewConnectionNotifier(rdb)
	ticketStore := redisinfra.NewWSTicketStore(rdb)

	// Auth use cases
	registerUC := coreauth.NewRegisterUseCase(userRepo)
	loginUC := coreauth.NewLoginUseCase(userRepo, sessionRepo)
	logoutUC := coreauth.NewLogoutUseCase(sessionRepo)
	issueTicketUC := coreauth.NewIssueWSTicketUseCase(sessionRepo, ticketStore)
	validateSessionUC := coreauth.NewValidateSessionUseCase(sessionRepo)

	// Handlers
	attachmentRepo := postgres.NewAttachmentRepository(db)
	handler := api.NewMessageHandler(repo, membershipRepo, attachmentRepo)
	roomContextHandler := api.NewRoomContextHandler(roomRepo, membershipRepo)
	messageContextHandler := api.NewMessageContextHandler(repo)
	roomMembershipWrite := api.NewRoomMembershipWriteHandler(producer, roomRepo, membershipRepo, userRepo)
	authHandler := api.NewAuthHandler(registerUC, loginUC, logoutUC, issueTicketUC, validateSessionUC, userRepo, connNotifier)
	mediaCfg := config.LoadMediaConfig()
	objectStorage := media.NewObjectStorageFromConfig(mediaCfg)
	type storageStatus interface {
		Configured() bool
	}
	if st, ok := objectStorage.(storageStatus); ok && st.Configured() {
		log.Logger.Info("MinIO connected",
			"bucket", mediaCfg.Bucket,
			"endpoint", mediaCfg.Endpoint,
			"upload_via_api", config.UseProxyUpload(),
			"minio_public_url", mediaCfg.PublicBaseURL,
		)
	} else {
		log.Logger.Warn("MinIO unavailable — media uploads disabled")
	}

	roomHandler := api.NewRoomHandler(roomRepo, membershipRepo, sessionRepo, objectStorage)
	unreadHandler := api.NewUnreadHandler(rdb, membershipRepo)
	mediaHandler := api.NewMediaHandler(roomRepo, membershipRepo, userRepo, objectStorage)

	// auth wraps a handler requiring a valid Bearer session token
	auth := func(h http.HandlerFunc) http.HandlerFunc {
		return api.AuthMiddleware(validateSessionUC, h)
	}

	// Routes — public
	http.HandleFunc("/register", authHandler.Register)
	http.HandleFunc("/login", authHandler.Login)
	http.HandleFunc("/logout", authHandler.Logout)
	http.HandleFunc("/ws-ticket", authHandler.IssueWSTicket)
	http.HandleFunc("/users/me", authHandler.Me)

	http.HandleFunc("/rooms/unread", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			auth(unreadHandler.List)(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/rooms/unread/clear", auth(unreadHandler.Clear))

	// /rooms requires authenticated session.
	http.HandleFunc("/rooms", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			auth(roomHandler.ListRooms)(w, r)
		case http.MethodPost:
			auth(roomHandler.CreateRoom)(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Routes — authenticated.
	http.HandleFunc("/room-members", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			auth(roomHandler.ListRoomMembers)(w, r)
		case http.MethodPost:
			auth(roomMembershipWrite.JoinRoom)(w, r)
		case http.MethodDelete:
			auth(roomMembershipWrite.LeaveRoom)(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/room-members/add", auth(roomMembershipWrite.AddMember))
	http.HandleFunc("/media/upload-request", auth(mediaHandler.UploadRequest))
	http.HandleFunc("/media/upload-put", auth(mediaHandler.UploadPut))
	http.HandleFunc("/media/upload-complete", auth(mediaHandler.UploadComplete))
	http.HandleFunc("/media/object", auth(mediaHandler.ServeObject))

	// Routes — mixed
	http.HandleFunc("/messages", auth(handler.GetMessages))

	// Routes — internal (called by gateway, no Bearer required)
	http.HandleFunc("/room-context", roomContextHandler.GetRoomContext)
	http.HandleFunc("/message-context", messageContextHandler.GetMessageContext)

	port := ":8081"
	log.Logger.Info("API running on port" + port)

	err = http.ListenAndServe(port, withCORS(http.DefaultServeMux))
	if err != nil {
		panic(err)
	}
}
