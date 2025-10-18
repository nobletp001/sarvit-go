package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Todo represents a post created by a user.
type Todo struct {
	ID        primitive.ObjectID   `bson:"_id,omitempty" json:"id"`
	AuthorID  primitive.ObjectID   `bson:"authorId" json:"authorId"`
	Title     string               `bson:"title" json:"title"`
	Body      string               `bson:"body" json:"body"`
	Likes     []primitive.ObjectID `bson:"likes" json:"likes"` // list of users who liked
	CreatedAt time.Time            `bson:"createdAt" json:"createdAt"`
	UpdatedAt time.Time            `bson:"updatedAt" json:"updatedAt"`
}
