// internal/transport/http/todo_handler.go
package http

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/nobletp001/sarvit/internal/service"
	"github.com/nobletp001/sarvit/internal/util"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type todoHandler struct {
	svc service.TodoService
}

func RegisterTodoRoutes(r fiber.Router, svc service.TodoService) {
	h := &todoHandler{svc: svc}

	// Protected routes (mounted under AuthMiddleware by caller)
	r.Post("/", h.create)
	r.Patch("/:id", h.update)
	r.Delete("/:id", h.delete)
	r.Post("/:id/like", h.like)
	r.Post("/:id/unlike", h.unlike)

	// Public read routes
	r.Get("/:id", h.getByID)
	r.Get("/author/:id", h.listByAuthor)
}

// ---------- Swagger DTOs ----------

// TodoCreateRequest is the request payload for creating a todo.
// swagger:model TodoCreateRequest
type TodoCreateRequest struct {
	// Title of the todo
	// example: Ship Swagger docs
	Title string `json:"title"`
	// Optional body/content
	// example: Add handlers annotations and push to Railway
	Body string `json:"body"`
}

// TodoUpdateRequest is the request payload for updating a todo (partial).
// swagger:model TodoUpdateRequest
type TodoUpdateRequest struct {
	// New title (optional)
	Title *string `json:"title"`
	// New body (optional)
	Body *string `json:"body"`
}

// ----------------------------------

// create creates a new todo.
//
// @Summary      Create todo
// @Tags         todos
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body  body      TodoCreateRequest  true  "Create payload"
// @Success      201   {object}  models.Todo
// @Failure      400   {object}  ErrorResponse
// @Router       /api/v1/todos [post]
func (h *todoHandler) create(c *fiber.Ctx) error {
	uid := c.Locals("uid").(primitive.ObjectID)

	var in TodoCreateRequest
	if err := c.BodyParser(&in); err != nil {
		return util.BadRequest(c, "invalid body", nil)
	}
	util.TrimAll(&in.Title, &in.Body)

	if in.Title == "" {
		return util.BadRequest(c, "title is required", nil)
	}

	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	t, err := h.svc.Create(ctx, uid, in.Title, in.Body)
	if err != nil {
		return util.BadRequest(c, err.Error(), nil)
	}
	return util.Created(c, "todo created", t)
}

// getByID returns a todo by ID.
//
// @Summary      Get todo by ID
// @Tags         todos
// @Produce      json
// @Param        id   path     string true "Todo ID (hex ObjectID)"
// @Success      200  {object} models.Todo
// @Failure      400  {object} ErrorResponse
// @Failure      404  {object} ErrorResponse
// @Router       /api/v1/todos/{id} [get]
func (h *todoHandler) getByID(c *fiber.Ctx) error {
	id, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return util.BadRequest(c, "bad id", nil)
	}

	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	t, err := h.svc.GetByID(ctx, id)
	if err != nil {
		return util.NotFound(c, err.Error())
	}
	return util.OK(c, "todo fetched", t)
}

// listByAuthor returns todos for a given author with pagination.
//
// @Summary      List todos by author
// @Tags         todos
// @Produce      json
// @Param        id     path   string true  "Author ID (hex ObjectID)"
// @Param        limit  query  int    false "Max items to return" default(20) minimum(1) maximum(100)
// @Param        skip   query  int    false "Items to skip (offset)" default(0) minimum(0)
// @Success      200    {array} models.Todo
// @Failure      400    {object} ErrorResponse
// @Router       /api/v1/todos/author/{id} [get]
func (h *todoHandler) listByAuthor(c *fiber.Ctx) error {
	authorID, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return util.BadRequest(c, "bad id", nil)
	}

	limitStr := c.Query("limit", "20")
	skipStr := c.Query("skip", "0")

	limit, err := strconv.ParseInt(limitStr, 10, 64)
	if err != nil {
		return util.BadRequest(c, "limit must be a number", nil)
	}
	skip, err := strconv.ParseInt(skipStr, 10, 64)
	if err != nil {
		return util.BadRequest(c, "skip must be a number", nil)
	}

	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	list, err := h.svc.ListByAuthor(ctx, authorID, limit, skip)
	if err != nil {
		return util.BadRequest(c, err.Error(), nil)
	}

	meta := fiber.Map{"limit": limit, "skip": skip, "count": len(list)}
	return util.WithMeta(c, "todos fetched", list, meta)
}

// update edits title/body of a todo.
//
// @Summary      Update todo
// @Tags         todos
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id    path      string             true  "Todo ID (hex ObjectID)"
// @Param        body  body      TodoUpdateRequest  true  "Update payload (at least one field)"
// @Success      204   "No Content"
// @Failure      400   {object}  ErrorResponse
// @Router       /api/v1/todos/{id} [patch]
func (h *todoHandler) update(c *fiber.Ctx) error {
	uid := c.Locals("uid").(primitive.ObjectID)

	id, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return util.BadRequest(c, "bad id", nil)
	}

	var in TodoUpdateRequest
	if err := c.BodyParser(&in); err != nil {
		return util.BadRequest(c, "invalid body", nil)
	}

	// Trim pointer fields safely
	if in.Title != nil {
		trimmed := strings.TrimSpace(*in.Title)
		in.Title = &trimmed
	}
	if in.Body != nil {
		trimmed := strings.TrimSpace(*in.Body)
		in.Body = &trimmed
	}

	// Require at least one non-empty field
	if (in.Title == nil || *in.Title == "") && (in.Body == nil || *in.Body == "") {
		return util.BadRequest(c, "nothing to update", nil)
	}

	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	if err := h.svc.Update(ctx, id, uid, in.Title, in.Body); err != nil {
		return util.BadRequest(c, err.Error(), nil)
	}
	return util.NoContent(c, "todo updated")
}

// delete removes a todo by ID.
//
// @Summary      Delete todo
// @Tags         todos
// @Security     BearerAuth
// @Produce      json
// @Param        id   path   string true "Todo ID (hex ObjectID)"
// @Success      204  "No Content"
// @Failure      400  {object} ErrorResponse
// @Router       /api/v1/todos/{id} [delete]
func (h *todoHandler) delete(c *fiber.Ctx) error {
	uid := c.Locals("uid").(primitive.ObjectID)

	id, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return util.BadRequest(c, "bad id", nil)
	}

	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	if err := h.svc.Delete(ctx, id, uid); err != nil {
		return util.BadRequest(c, err.Error(), nil)
	}
	return util.NoContent(c, "todo deleted")
}

// like toggles a like on.
//
// @Summary      Like todo
// @Tags         todos
// @Security     BearerAuth
// @Produce      json
// @Param        id   path   string true "Todo ID (hex ObjectID)"
// @Success      204  "No Content"
// @Failure      400  {object} ErrorResponse
// @Router       /api/v1/todos/{id}/like [post]
func (h *todoHandler) like(c *fiber.Ctx) error {
	uid := c.Locals("uid").(primitive.ObjectID)

	id, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return util.BadRequest(c, "bad id", nil)
	}

	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	if err := h.svc.ToggleLike(ctx, id, uid, true); err != nil {
		return util.BadRequest(c, err.Error(), nil)
	}
	return util.NoContent(c, "liked")
}

// unlike removes a like.
//
// @Summary      Unlike todo
// @Tags         todos
// @Security     BearerAuth
// @Produce      json
// @Param        id   path   string true "Todo ID (hex ObjectID)"
// @Success      204  "No Content"
// @Failure      400  {object} ErrorResponse
// @Router       /api/v1/todos/{id}/unlike [post]
func (h *todoHandler) unlike(c *fiber.Ctx) error {
	uid := c.Locals("uid").(primitive.ObjectID)

	id, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return util.BadRequest(c, "bad id", nil)
	}

	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	if err := h.svc.ToggleLike(ctx, id, uid, false); err != nil {
		return util.BadRequest(c, err.Error(), nil)
	}
	return util.NoContent(c, "unliked")
}
