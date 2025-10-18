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

var ErrTodoNotFound = errors.New("todo not found")

type TodoRepository interface {
	Create(ctx context.Context, t *models.Todo) error
	GetByID(ctx context.Context, id primitive.ObjectID) (*models.Todo, error)
	ListByAuthor(ctx context.Context, authorID primitive.ObjectID, limit, skip int64) ([]models.Todo, error)
	Update(ctx context.Context, id primitive.ObjectID, update bson.M) error
	Delete(ctx context.Context, id primitive.ObjectID) error
	ToggleLike(ctx context.Context, id, userID primitive.ObjectID, like bool) error
}

type todoRepo struct {
	col *mongo.Collection
}

func NewTodoRepository(client *mongo.Client, dbName string) (TodoRepository, error) {
	col := client.Database(dbName).Collection("todos")
	// Useful indexes
	_, err := col.Indexes().CreateMany(context.Background(), []mongo.IndexModel{
		{Keys: bson.D{{Key: "authorId", Value: 1}, {Key: "createdAt", Value: -1}}},
		{Keys: bson.D{{Key: "createdAt", Value: -1}}},
	})
	if err != nil {
		return nil, err
	}
	return &todoRepo{col: col}, nil
}

func (r *todoRepo) Create(ctx context.Context, t *models.Todo) error {
	now := time.Now().UTC()
	if t.CreatedAt.IsZero() {
		t.CreatedAt = now
	}
	t.UpdatedAt = now
	if t.Likes == nil {
		t.Likes = []primitive.ObjectID{}
	}
	_, err := r.col.InsertOne(ctx, t)
	return err
}

func (r *todoRepo) GetByID(ctx context.Context, id primitive.ObjectID) (*models.Todo, error) {
	var out models.Todo
	err := r.col.FindOne(ctx, bson.M{"_id": id}).Decode(&out)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrTodoNotFound
		}
		return nil, err
	}
	return &out, nil
}

func (r *todoRepo) ListByAuthor(ctx context.Context, authorID primitive.ObjectID, limit, skip int64) ([]models.Todo, error) {
	opts := options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}})
	if limit > 0 {
		opts.SetLimit(limit)
	}
	if skip > 0 {
		opts.SetSkip(skip)
	}
	cur, err := r.col.Find(ctx, bson.M{"authorId": authorID}, opts)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var list []models.Todo
	for cur.Next(ctx) {
		var t models.Todo
		if err := cur.Decode(&t); err != nil {
			return nil, err
		}
		list = append(list, t)
	}
	return list, cur.Err()
}

func (r *todoRepo) Update(ctx context.Context, id primitive.ObjectID, update bson.M) error {
	// Always bump updatedAt
	set := bson.M{"updatedAt": time.Now().UTC()}
	if u, ok := update["$set"].(bson.M); ok {
		for k, v := range u {
			set[k] = v
		}
		update["$set"] = set
	} else {
		update["$set"] = set
	}

	res, err := r.col.UpdateByID(ctx, id, update)
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return ErrTodoNotFound
	}
	return nil
}

func (r *todoRepo) Delete(ctx context.Context, id primitive.ObjectID) error {
	res, err := r.col.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return ErrTodoNotFound
	}
	return nil
}

func (r *todoRepo) ToggleLike(ctx context.Context, id, userID primitive.ObjectID, like bool) error {
	var update bson.M
	if like {
		update = bson.M{"$addToSet": bson.M{"likes": userID}}
	} else {
		update = bson.M{"$pull": bson.M{"likes": userID}}
	}
	res, err := r.col.UpdateByID(ctx, id, update)
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return ErrTodoNotFound
	}
	return nil
}
