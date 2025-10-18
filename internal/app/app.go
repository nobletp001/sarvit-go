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
	"go.mongodb.org/mongo-driver/mongo" // ✅ add this import
)

func Run() error {
	cfg := config.Load()

	// =============================
	// STEP 1: SELECT DATABASE DRIVER
	// =============================
	dbDriver := cfg.DBDriver
	if dbDriver == "" {
		dbDriver = "mongo" // default
	}
	log.Printf("🧠 Using database driver: %s", dbDriver)

	var (
		client      *mongo.Client // ✅ corrected type
		userRepo    repository.UserRepository
		todoRepo    repository.TodoRepository
		commentRepo repository.CommentRepository
	)

	switch dbDriver {
	case "mongo":
		// --- Connect to Mongo ---
		mongoClient, err := db.Connect(cfg)
		if err != nil {
			log.Printf("❌ MongoDB connect/ping failed: %v", err)
			return err
		}
		defer func() { _ = mongoClient.Disconnect(context.Background()) }()
		log.Printf("✅ Mongo connected (db=%s)", cfg.MongoDB)

		client = mongoClient // assign handle

		// --- Initialize Mongo Repositories ---
		userRepo, err = repository.NewUserRepository(client, cfg.MongoDB)
		if err != nil {
			return err
		}
		todoRepo, err = repository.NewTodoRepository(client, cfg.MongoDB)
		if err != nil {
			return err
		}
		commentRepo, err = repository.NewCommentRepository(client, cfg.MongoDB)
		if err != nil {
			return err
		}

	default:
		log.Fatalf("❌ Unsupported DB driver: %s (only 'mongo' supported now)", dbDriver)
	}

	// =============================
	// STEP 2: SERVICES
	// =============================
	authSvc := service.NewAuthService(cfg, userRepo)
	todoSvc := service.NewTodoService(todoRepo)
	commentSvc := service.NewCommentService(commentRepo, todoRepo)

	// =============================
	// STEP 3: HTTP SERVER
	// =============================
	app := httptransport.NewServer(cfg, authSvc, todoSvc, commentSvc)

	// Port fallback
	port := cfg.Port
	if port == "" {
		port = "3000"
	}
	addr := ":" + port
	log.Printf("🚀 Server starting on http://localhost%s", addr)

	// =============================
	// STEP 4: START SERVER & HANDLE SHUTDOWN
	// =============================
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
