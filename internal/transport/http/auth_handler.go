package http

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/nobletp001/sarvit/internal/service"
	"github.com/nobletp001/sarvit/internal/util"
)

type authHandler struct {
	svc service.AuthService
}

func RegisterAuthRoutes(r fiber.Router, svc service.AuthService) {
	h := &authHandler{svc: svc}
	r.Post("/signup", h.signup)
	r.Post("/login", h.login)
}

// --- Request/Response DTOs ---

type signupReq struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type authRes struct {
	Token string      `json:"token"`
	User  interface{} `json:"user"`
}

// --- Handlers ---

func (h *authHandler) signup(c *fiber.Ctx) error {
	var in signupReq
	if err := c.BodyParser(&in); err != nil {
		return util.BadRequest(c, "invalid body", nil)
	}
	util.TrimAll(&in.Name, &in.Email, &in.Password)

	// Specific validation messages
	if in.Name == "" {
		return util.BadRequest(c, "name is required", nil)
	}
	if in.Email == "" {
		return util.BadRequest(c, "email is required", nil)
	}
	if in.Password == "" {
		return util.BadRequest(c, "password is required", nil)
	}

	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	u, tok, err := h.svc.Signup(ctx, in.Name, in.Email, in.Password)
	if err != nil {
		switch err {
		case service.ErrEmailExists:
			return c.Status(fiber.StatusConflict).JSON(util.APIResponse{
				Status:  "error",
				Message: "email already in use",
				Data:    nil,
			})
		default:
			return util.BadRequest(c, err.Error(), nil)
		}
	}

	return util.Created(c, "user registered successfully", authRes{Token: tok, User: u})
}

func (h *authHandler) login(c *fiber.Ctx) error {
	var in loginReq
	if err := c.BodyParser(&in); err != nil {
		return util.BadRequest(c, "invalid body", nil)
	}
	util.TrimAll(&in.Email, &in.Password)

	// Specific validation messages
	if in.Email == "" {
		return util.BadRequest(c, "email is required", nil)
	}
	if in.Password == "" {
		return util.BadRequest(c, "password is required", nil)
	}

	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	u, tok, err := h.svc.Login(ctx, in.Email, in.Password)
	if err != nil {
		if err == service.ErrInvalidCredentials {
			// Explicit message as requested
			return c.Status(fiber.StatusUnauthorized).JSON(util.APIResponse{
				Status:  "error",
				Message: "password is not correct",
				Data:    nil,
			})
		}
		return util.Unauthorized(c, err.Error())
	}

	return util.OK(c, "login successful", authRes{Token: tok, User: u})
}
