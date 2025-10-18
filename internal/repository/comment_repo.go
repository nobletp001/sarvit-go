package repository

import (
	"context"
	"errors"
	"time"

	"github.com/nobletp001/sarvit/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var ErrCommentNotFound = errors.New("comment not found")

type CommentRepository interface {
	Create(ctx context.Context, c *models.Comment) error
	ListByTodo(ctx context.Context, todoID primitive.ObjectID, limit, skip int64) ([]models.Comment, error)
	Delete(ctx context.Context, id primitive.ObjectID) error
}

type commentRepo struct {
	col *mongo.Collection
}

func NewCommentRepository(client *mongo.Client, dbName string) (CommentRepository, error) {
	col := client.Database(dbName).Collection("comments")
	_, err := col.Indexes().CreateMany(context.Background(), []mongo.IndexModel{
		{Keys: bson.D{{Key: "todoId", Value: 1}, {Key: "createdAt", Value: -1}}},
		{Keys: bson.D{{Key: "createdAt", Value: -1}}},
	})
	if err != nil {
		return nil, err
	}
	return &commentRepo{col: col}, nil
}

func (r *commentRepo) Create(ctx context.Context, c *models.Comment) error {
	if c.CreatedAt.IsZero() {
		c.CreatedAt = time.Now().UTC()
	}
	_, err := r.col.InsertOne(ctx, c)
	return err
}

func (r *commentRepo) ListByTodo(ctx context.Context, todoID primitive.ObjectID, limit, skip int64) ([]models.Comment, error) {
	opts := options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}})
	if limit > 0 {
		opts.SetLimit(limit)
	}
	if skip > 0 {
		opts.SetSkip(skip)
	}
	cur, err := r.col.Find(ctx, bson.M{"todoId": todoID}, opts)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var list []models.Comment
	for cur.Next(ctx) {
		var c models.Comment
		if err := cur.Decode(&c); err != nil {
			return nil, err
		}
		list = append(list, c)
	}
	return list, cur.Err()
}

func (r *commentRepo) Delete(ctx context.Context, id primitive.ObjectID) error {
	res, err := r.col.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return ErrCommentNotFound
	}
	return nil
}
