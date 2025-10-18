package util

import (
	"strings"

	"github.com/gofiber/fiber/v2"
)

// APIResponse is the unified response envelope.
type APIResponse struct {
	Status  string      `json:"status"`         // "success" | "error"
	Message string      `json:"message"`        // human-friendly message
	Data    interface{} `json:"data"`           // payload or null
	Meta    interface{} `json:"meta,omitempty"` // optional: pagination, etc.
}

// JSON sends a custom envelope.
func JSON(c *fiber.Ctx, code int, status, message string, data, meta interface{}) error {
	return c.Status(code).JSON(APIResponse{
		Status:  status,
		Message: message,
		Data:    data,
		Meta:    meta,
	})
}

// --- Success helpers ---

func OK(c *fiber.Ctx, message string, data interface{}) error {
	return JSON(c, fiber.StatusOK, "success", message, data, nil)
}

func Created(c *fiber.Ctx, message string, data interface{}) error {
	return JSON(c, fiber.StatusCreated, "success", message, data, nil)
}

func NoContent(c *fiber.Ctx, message string) error {
	// Keep envelope uniform even for 204-like semantics
	return JSON(c, fiber.StatusOK, "success", message, nil, nil)
}

// WithMeta is handy for list endpoints (pagination, counts, etc.)
func WithMeta(c *fiber.Ctx, message string, data, meta interface{}) error {
	return JSON(c, fiber.StatusOK, "success", message, data, meta)
}

// --- Error helpers ---

func BadRequest(c *fiber.Ctx, message string, data interface{}) error {
	return JSON(c, fiber.StatusBadRequest, "error", message, data, nil)
}

func Unauthorized(c *fiber.Ctx, message string) error {
	return JSON(c, fiber.StatusUnauthorized, "error", message, nil, nil)
}

func Forbidden(c *fiber.Ctx, message string) error {
	return JSON(c, fiber.StatusForbidden, "error", message, nil, nil)
}

func NotFound(c *fiber.Ctx, message string) error {
	return JSON(c, fiber.StatusNotFound, "error", message, nil, nil)
}

func Internal(c *fiber.Ctx, message string) error {
	return JSON(c, fiber.StatusInternalServerError, "error", message, nil, nil)
}

// ErrorMiddleware converts any returned error into our JSON envelope.
// Place it BEFORE recover/logger so Fiber pipes errors to it.
func ErrorMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if err := c.Next(); err != nil {
			code := fiber.StatusInternalServerError
			msg := err.Error()
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
				msg = e.Message
			}
			return JSON(c, code, "error", msg, nil, nil)
		}
		return nil
	}
}

func TrimAll(ss ...*string) {
	for _, s := range ss {
		if s != nil {
			*s = strings.TrimSpace(*s)
		}
	}
}
