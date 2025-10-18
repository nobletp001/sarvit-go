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

	// Protected routes (this group is mounted under AuthMiddleware in router.go)
	r.Post("/", h.create)
	r.Patch("/:id", h.update)
	r.Delete("/:id", h.delete)
	r.Post("/:id/like", h.like)
	r.Post("/:id/unlike", h.unlike)

	// Public read routes (still accessible; if you want them protected, move above)
	r.Get("/:id", h.getByID)
	r.Get("/author/:id", h.listByAuthor)
}

type todoCreateReq struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

func (h *todoHandler) create(c *fiber.Ctx) error {
	uid := c.Locals("uid").(primitive.ObjectID)

	var in todoCreateReq
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

type todoUpdateReq struct {
	Title *string `json:"title"`
	Body  *string `json:"body"`
}

func (h *todoHandler) update(c *fiber.Ctx) error {
	uid := c.Locals("uid").(primitive.ObjectID)

	id, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return util.BadRequest(c, "bad id", nil)
	}

	var in todoUpdateReq
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
