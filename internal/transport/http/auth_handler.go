package http

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/nobletp001/sarvit/internal/service"
	"github.com/nobletp001/sarvit/internal/util"
)

// ---------- Swagger DTOs ----------

// swagger:model SignupRequest
type SignupRequest struct {
	// The user's full name
	// example: Temitope Joshua
	Name string `json:"name"`
	// The user's email address
	// example: joshua@example.com
	Email string `json:"email"`
	// The user's password
	// example: StrongPass#123
	Password string `json:"password"`
}

// swagger:model LoginRequest
type LoginRequest struct {
	// The user's email address
	// example: joshua@example.com
	Email string `json:"email"`
	// The user's password
	// example: StrongPass#123
	Password string `json:"password"`
}

// swagger:model AuthResponse
type AuthResponse struct {
	// The JWT bearer token
	// example: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
	Token string `json:"token"`
	// The authenticated user object
	User interface{} `json:"user"`
}

// swagger:model ErrorResponse
type ErrorResponse struct {
	// always "error" on failures
	// example: error
	Status string `json:"status"`
	// human-readable message
	// example: email already in use
	Message string `json:"message"`
	// optional data payload (usually null on error)
	Data interface{} `json:"data"`
}

// ----------------------------------

type authHandler struct {
	svc service.AuthService
}

func RegisterAuthRoutes(r fiber.Router, svc service.AuthService) {
	h := &authHandler{svc: svc}
	r.Post("/signup", h.signup)
	r.Post("/login", h.login)
}

// signup registers a new user.
//
// @Summary      Register a new user
// @Description  Creates a user account and returns a JWT + user
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      SignupRequest  true  "Signup payload"
// @Success      201   {object}  AuthResponse
// @Failure      400   {object}  ErrorResponse
// @Failure      409   {object}  ErrorResponse  "email already in use"
// @Router       /api/v1/auth/signup [post]
func (h *authHandler) signup(c *fiber.Ctx) error {
	var in SignupRequest
	if err := c.BodyParser(&in); err != nil {
		return util.BadRequest(c, "invalid body", nil)
	}
	util.TrimAll(&in.Name, &in.Email, &in.Password)

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

	return util.Created(c, "user registered successfully", AuthResponse{Token: tok, User: u})
}

// login authenticates a user and returns a JWT.
//
// @Summary      Login
// @Description  Authenticates a user with email/password and returns a JWT + user
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      LoginRequest  true  "Login credentials"
// @Success      200   {object}  AuthResponse
// @Failure      400   {object}  ErrorResponse
// @Failure      401   {object}  ErrorResponse  "password is not correct"
// @Router       /api/v1/auth/login [post]
func (h *authHandler) login(c *fiber.Ctx) error {
	var in LoginRequest
	if err := c.BodyParser(&in); err != nil {
		return util.BadRequest(c, "invalid body", nil)
	}
	util.TrimAll(&in.Email, &in.Password)

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
			return c.Status(fiber.StatusUnauthorized).JSON(util.APIResponse{
				Status:  "error",
				Message: "password is not correct",
				Data:    nil,
			})
		}
		return util.Unauthorized(c, err.Error())
	}

	return util.OK(c, "login successful", AuthResponse{Token: tok, User: u})
}
