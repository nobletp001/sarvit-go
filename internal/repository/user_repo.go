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

var ErrUserNotFound = errors.New("user not found")

type UserRepository interface {
	Create(ctx context.Context, u *models.User) error
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	FindByID(ctx context.Context, id primitive.ObjectID) (*models.User, error)
}

type userRepo struct {
	col *mongo.Collection
}

func NewUserRepository(client *mongo.Client, dbName string) (UserRepository, error) {
	col := client.Database(dbName).Collection("users")
	// Unique index on email
	_, err := col.Indexes().CreateOne(context.Background(), mongo.IndexModel{
		Keys:    bson.D{{Key: "email", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		return nil, err
	}
	return &userRepo{col: col}, nil
}

func (r *userRepo) Create(ctx context.Context, u *models.User) error {
	u.ID = primitive.NilObjectID // let Mongo assign if empty
	if u.CreatedAt.IsZero() {
		u.CreatedAt = time.Now().UTC()
	}
	_, err := r.col.InsertOne(ctx, u)
	return err
}

func (r *userRepo) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	var out models.User
	err := r.col.FindOne(ctx, bson.M{"email": email}).Decode(&out)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &out, nil
}

func (r *userRepo) FindByID(ctx context.Context, id primitive.ObjectID) (*models.User, error) {
	var out models.User
	err := r.col.FindOne(ctx, bson.M{"_id": id}).Decode(&out)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &out, nil
}
