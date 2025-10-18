package service

import (
	"context"
	"errors"
	"time"

	"github.com/nobletp001/sarvit/internal/models"
	"github.com/nobletp001/sarvit/internal/repository"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	ErrUnauthorized = errors.New("not allowed")
)

type TodoService interface {
	Create(ctx context.Context, authorID primitive.ObjectID, title, body string) (*models.Todo, error)
	GetByID(ctx context.Context, id primitive.ObjectID) (*models.Todo, error)
	ListByAuthor(ctx context.Context, authorID primitive.ObjectID, limit, skip int64) ([]models.Todo, error)
	Update(ctx context.Context, id, requestorID primitive.ObjectID, title, body *string) error
	Delete(ctx context.Context, id, requestorID primitive.ObjectID) error
	ToggleLike(ctx context.Context, id, userID primitive.ObjectID, like bool) error
}

type todoService struct {
	todos repository.TodoRepository
}

func NewTodoService(todos repository.TodoRepository) TodoService {
	return &todoService{todos: todos}
}

func (s *todoService) Create(ctx context.Context, authorID primitive.ObjectID, title, body string) (*models.Todo, error) {
	t := &models.Todo{
		ID:        primitive.NewObjectID(),
		AuthorID:  authorID,
		Title:     title,
		Body:      body,
		Likes:     []primitive.ObjectID{},
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	if err := s.todos.Create(ctx, t); err != nil {
		return nil, err
	}
	return t, nil
}

func (s *todoService) GetByID(ctx context.Context, id primitive.ObjectID) (*models.Todo, error) {
	return s.todos.GetByID(ctx, id)
}

func (s *todoService) ListByAuthor(ctx context.Context, authorID primitive.ObjectID, limit, skip int64) ([]models.Todo, error) {
	return s.todos.ListByAuthor(ctx, authorID, limit, skip)
}

func (s *todoService) Update(ctx context.Context, id, requestorID primitive.ObjectID, title, body *string) error {
	current, err := s.todos.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if current.AuthorID != requestorID {
		return ErrUnauthorized
	}
	set := bson.M{}
	if title != nil {
		set["title"] = *title
	}
	if body != nil {
		set["body"] = *body
	}
	return s.todos.Update(ctx, id, bson.M{"$set": set})
}

func (s *todoService) Delete(ctx context.Context, id, requestorID primitive.ObjectID) error {
	current, err := s.todos.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if current.AuthorID != requestorID {
		return ErrUnauthorized
	}
	return s.todos.Delete(ctx, id)
}

func (s *todoService) ToggleLike(ctx context.Context, id, userID primitive.ObjectID, like bool) error {
	return s.todos.ToggleLike(ctx, id, userID, like)
}
