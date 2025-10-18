package http

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"

	"github.com/nobletp001/sarvit/internal/config"
	"github.com/nobletp001/sarvit/internal/middleware" // ✅ new import
	"github.com/nobletp001/sarvit/internal/service"
	"github.com/nobletp001/sarvit/internal/util"
)

// NewServer builds and configures the Fiber HTTP server.
func NewServer(
	cfg config.Config,
	authSvc service.AuthService,
	todoSvc service.TodoService,
	commentSvc service.CommentService,
) *fiber.App {
	app := fiber.New()

	// Global middleware stack
	app.Use(util.ErrorMiddleware()) // unified JSON error responses
	app.Use(logger.New())           // request logging
	app.Use(recover.New())          // recover from panics
	app.Use(cors.New())             // allow cross-origin requests

	// --- Health route ---
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "ok",
			"env":    cfg.AppEnv,
			"db":     cfg.MongoDB,
			"driver": cfg.DBDriver,
		})
	})

	// --- Versioned API group ---
	api := app.Group("/api/v1")

	// --- Public routes ---
	RegisterAuthRoutes(api.Group("/auth"), authSvc)

	// --- Protected routes (JWT required) ---
	protected := api.Group("", middleware.AuthMiddleware(authSvc))
	RegisterTodoRoutes(protected.Group("/todos"), todoSvc)
	RegisterCommentRoutes(protected.Group("/comments"), commentSvc)

	return app
}
