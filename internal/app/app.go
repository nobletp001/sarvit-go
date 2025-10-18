package app

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nobletp001/sarvit/internal/config"
	"github.com/nobletp001/sarvit/internal/db"
	"github.com/nobletp001/sarvit/internal/repository"
	"github.com/nobletp001/sarvit/internal/service"
	httptransport "github.com/nobletp001/sarvit/internal/transport/http"
	"go.mongodb.org/mongo-driver/mongo"
)

func Run() error {
	cfg := config.Load()

	// ========== DB DRIVER ==========
	dbDriver := cfg.DBDriver
	if dbDriver == "" {
		dbDriver = "mongo"
	}
	log.Printf("🧠 Using database driver: %s", dbDriver)

	var (
		client      *mongo.Client
		userRepo    repository.UserRepository
		todoRepo    repository.TodoRepository
		commentRepo repository.CommentRepository
	)

	switch dbDriver {
	case "mongo":
		mongoClient, err := db.Connect(cfg)
		if err != nil {
			log.Printf("❌ MongoDB connect/ping failed: %v", err)
			return err
		}
		defer func() { _ = mongoClient.Disconnect(context.Background()) }()
		log.Printf("✅ Mongo connected (db=%s)", cfg.MongoDB)

		client = mongoClient

		var errRepo error
		userRepo, errRepo = repository.NewUserRepository(client, cfg.MongoDB)
		if errRepo != nil {
			return errRepo
		}
		todoRepo, errRepo = repository.NewTodoRepository(client, cfg.MongoDB)
		if errRepo != nil {
			return errRepo
		}
		commentRepo, errRepo = repository.NewCommentRepository(client, cfg.MongoDB)
		if errRepo != nil {
			return errRepo
		}
	default:
		log.Fatalf("❌ Unsupported DB driver: %s (only 'mongo' supported)", dbDriver)
	}

	// ========== SERVICES ==========
	authSvc := service.NewAuthService(cfg, userRepo)
	todoSvc := service.NewTodoService(todoRepo)
	commentSvc := service.NewCommentService(commentRepo, todoRepo)

	// ========== HTTP SERVER ==========
	app := httptransport.NewServer(cfg, authSvc, todoSvc, commentSvc)

	// ---- LISTEN ON RAILWAY PORT (fallbacks for local) ----
	port := os.Getenv("PORT") // provided by Railway
	if port == "" {
		if cfg.Port != "" {
			port = cfg.Port
		} else {
			port = "3000"
		}
	}
	addr := ":" + port
	log.Printf("🚀 Server starting on %s (RAILWAY PORT=%s)", addr, os.Getenv("PORT"))

	// ========== START & SHUTDOWN ==========
	errCh := make(chan error, 1)
	go func() {
		if err := app.Listen(addr); err != nil {
			errCh <- err
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-quit:
		log.Printf("🛑 Caught signal: %s — shutting down...", sig)
	case srvErr := <-errCh:
		if srvErr != nil && !errors.Is(srvErr, context.Canceled) {
			log.Printf("❌ Server error: %v", srvErr)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := app.ShutdownWithContext(ctx); err != nil {
		log.Printf("⚠️ Fiber shutdown error: %v", err)
	} else {
		log.Println("✅ Server stopped cleanly")
	}
	return nil
}
