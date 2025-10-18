package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/nobletp001/sarvit/internal/service"
	"github.com/nobletp001/sarvit/internal/util"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// AuthMiddleware verifies JWT and sets Locals("uid") = primitive.ObjectID.
func AuthMiddleware(auth service.AuthService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := strings.TrimSpace(c.Get("Authorization"))
		if authHeader == "" {
			return util.Unauthorized(c, "authorization header missing")
		}

		if !strings.HasPrefix(authHeader, "Bearer ") {
			return util.Unauthorized(c, "invalid authorization format, expected 'Bearer <token>'")
		}

		token := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
		if token == "" {
			return util.Unauthorized(c, "empty bearer token")
		}

		uid, err := auth.ParseUserIDFromToken(token)
		if err != nil || uid == primitive.NilObjectID {
			return util.Unauthorized(c, "invalid or expired token")
		}

		// Store user ID in context for handlers to access
		c.Locals("uid", uid)
		return c.Next()
	}
}
