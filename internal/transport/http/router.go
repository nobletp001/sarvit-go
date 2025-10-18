// internal/transport/http/server.go
package http

// Package http Sarvit API
//
// @title           Sarvit Go API
// @version         1.0
// @description     REST API (Fiber + MongoDB)
// @schemes         https http
// @BasePath        /
//
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Enter "Bearer <token>"

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"

	"github.com/gofiber/swagger"
	_ "github.com/nobletp001/sarvit/docs"

	"github.com/nobletp001/sarvit/internal/config"
	"github.com/nobletp001/sarvit/internal/middleware"
	"github.com/nobletp001/sarvit/internal/service"
	"github.com/nobletp001/sarvit/internal/util"
)

// healthHandler returns basic app/env/db info.
//
// @Summary      Health
// @Tags         meta
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Router       /health [get]
func healthHandler(cfg config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "ok",
			"env":    cfg.AppEnv,
			"db":     cfg.MongoDB,
			"driver": cfg.DBDriver,
		})
	}
}

// NewServer builds and configures the Fiber HTTP server.
func NewServer(
	cfg config.Config,
	authSvc service.AuthService,
	todoSvc service.TodoService,
	commentSvc service.CommentService,
) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName: "Sarvit API",
	})

	// ---------- Global middleware ----------
	app.Use(util.ErrorMiddleware()) // unified JSON error responses
	app.Use(logger.New())           // request logging
	app.Use(recover.New())          // panic recovery
	app.Use(helmet.New())           // security headers
	app.Use(compress.New())         // gzip compression

	// Rate limit to prevent abuse
	app.Use(limiter.New(limiter.Config{
		Max:        100,             // requests
		Expiration: 1 * time.Minute, // per IP per minute
	}))

	// Allow CORS for specific origins or all
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*", // change to your frontend: e.g., "https://app.sarvit.com"
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowMethods: "GET,POST,PATCH,DELETE,OPTIONS",
	}))

	// ---------- Routes ----------
	app.Get("/health", healthHandler(cfg))

	// Swagger docs (disable in production)
	if cfg.AppEnv != "production" {
		app.Get("/swagger/*", swagger.New(swagger.Config{}))
	}

	api := app.Group("/api/v1")

	// Public routes
	RegisterAuthRoutes(api.Group("/auth"), authSvc)

	// Protected routes (JWT required)
	protected := api.Group("", middleware.AuthMiddleware(authSvc))
	RegisterTodoRoutes(protected.Group("/todos"), todoSvc)
	RegisterCommentRoutes(protected.Group("/comments"), commentSvc)

	return app
}
