package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Comment represents a user comment on a todo.
type Comment struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	TodoID    primitive.ObjectID `bson:"todoId" json:"todoId"`
	AuthorID  primitive.ObjectID `bson:"authorId" json:"authorId"`
	Body      string             `bson:"body" json:"body"`
	CreatedAt time.Time          `bson:"createdAt" json:"createdAt"`
}
