// internal/transport/http/comment_handler.go
package http

import (
	"context"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/nobletp001/sarvit/internal/service"
	"github.com/nobletp001/sarvit/internal/util"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type commentHandler struct {
	svc service.CommentService
}

// RegisterCommentRoutes mounts protected comment routes.
// All routes require JWT (AuthMiddleware applied by caller).
func RegisterCommentRoutes(r fiber.Router, svc service.CommentService) {
	h := &commentHandler{svc: svc}
	// POST   /comments/:todoId   -> add a comment to a todo
	// GET    /comments/:todoId   -> list comments for a todo
	// DELETE /comments/:id       -> delete a comment by id (owner only)
	r.Post("/:todoId", h.add)
	r.Get("/:todoId", h.list)
	r.Delete("/:id", h.delete)
}

// ---------- Swagger DTOs ----------

// CommentRequest is the request body for creating a comment.
// swagger:model CommentRequest
type CommentRequest struct {
	// The body text of the comment
	// example: This helped a lot—thanks!
	Body string `json:"body"`
}

// ----------------------------------

// add creates a new comment on a todo.
//
// @Summary      Add a comment to a todo
// @Tags         comments
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        todoId  path      string          true  "Todo ID (hex ObjectID)"
// @Param        body    body      CommentRequest  true  "Comment payload"
// @Success      201     {object}  models.Comment
// @Failure      400     {object}  ErrorResponse
// @Router       /api/v1/comments/{todoId} [post]
func (h *commentHandler) add(c *fiber.Ctx) error {
	uid := c.Locals("uid").(primitive.ObjectID)

	todoID, err := primitive.ObjectIDFromHex(c.Params("todoId"))
	if err != nil {
		return util.BadRequest(c, "bad todo id", nil)
	}

	var in CommentRequest
	if err := c.BodyParser(&in); err != nil {
		return util.BadRequest(c, "invalid body", nil)
	}
	util.TrimAll(&in.Body)
	if in.Body == "" {
		return util.BadRequest(c, "body is required", nil)
	}

	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	com, err := h.svc.Add(ctx, todoID, uid, in.Body)
	if err != nil {
		return util.BadRequest(c, err.Error(), nil)
	}
	return util.Created(c, "comment created", com)
}

// list returns paginated comments for a todo.
//
// @Summary      List comments for a todo
// @Tags         comments
// @Security     BearerAuth
// @Produce      json
// @Param        todoId  path   string true  "Todo ID (hex ObjectID)"
// @Param        limit   query  int    false "Max items to return" default(20) minimum(1) maximum(100)
// @Param        skip    query  int    false "Items to skip (offset)" default(0) minimum(0)
// @Success      200     {array}  models.Comment
// @Failure      400     {object} ErrorResponse
// @Router       /api/v1/comments/{todoId} [get]
func (h *commentHandler) list(c *fiber.Ctx) error {
	todoID, err := primitive.ObjectIDFromHex(c.Params("todoId"))
	if err != nil {
		return util.BadRequest(c, "bad todo id", nil)
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

	list, err := h.svc.ListByTodo(ctx, todoID, limit, skip)
	if err != nil {
		return util.BadRequest(c, err.Error(), nil)
	}

	meta := fiber.Map{"limit": limit, "skip": skip, "count": len(list)}
	return util.WithMeta(c, "comments fetched", list, meta)
}

// delete removes a comment by ID (owner only).
//
// @Summary      Delete a comment
// @Tags         comments
// @Security     BearerAuth
// @Produce      json
// @Param        id   path   string true "Comment ID (hex ObjectID)"
// @Success      204  "No Content"
// @Failure      400  {object} ErrorResponse
// @Router       /api/v1/comments/{id} [delete]
func (h *commentHandler) delete(c *fiber.Ctx) error {
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
	return util.NoContent(c, "comment deleted")
}
