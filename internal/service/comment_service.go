package service

import (
	"context"
	"errors"
	"time"

	"github.com/nobletp001/sarvit/internal/models"
	"github.com/nobletp001/sarvit/internal/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CommentService interface {
	Add(ctx context.Context, todoID, authorID primitive.ObjectID, body string) (*models.Comment, error)
	ListByTodo(ctx context.Context, todoID primitive.ObjectID, limit, skip int64) ([]models.Comment, error)
	Delete(ctx context.Context, id, requestorID primitive.ObjectID) error
}

type commentService struct {
	comments repository.CommentRepository
	todos    repository.TodoRepository
}

func NewCommentService(comments repository.CommentRepository, todos repository.TodoRepository) CommentService {
	return &commentService{comments: comments, todos: todos}
}

func (s *commentService) Add(ctx context.Context, todoID, authorID primitive.ObjectID, body string) (*models.Comment, error) {
	// ensure todo exists
	if _, err := s.todos.GetByID(ctx, todoID); err != nil {
		return nil, err
	}

	c := &models.Comment{
		ID:        primitive.NewObjectID(),
		TodoID:    todoID,
		AuthorID:  authorID,
		Body:      body,
		CreatedAt: time.Now().UTC(),
	}
	if err := s.comments.Create(ctx, c); err != nil {
		return nil, err
	}
	return c, nil
}

func (s *commentService) ListByTodo(ctx context.Context, todoID primitive.ObjectID, limit, skip int64) ([]models.Comment, error) {
	return s.comments.ListByTodo(ctx, todoID, limit, skip)
}

func (s *commentService) Delete(ctx context.Context, id, requestorID primitive.ObjectID) error {
	// We need to fetch comment to check ownership.
	list, err := s.comments.ListByTodo(ctx, primitive.NilObjectID, 0, 0) // placeholder: no direct GetByID in repo
	if err != nil {
		return err
	}
	// In practice, add CommentRepository.GetByID to avoid scanning.
	var found *models.Comment
	for i := range list {
		if list[i].ID == id {
			found = &list[i]
			break
		}
	}
	if found == nil {
		return errors.New("comment not found")
	}
	if found.AuthorID != requestorID {
		return ErrUnauthorized
	}
	return s.comments.Delete(ctx, id)
}
