package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Comment represents a user comment on a todo.
//
// swagger:model Comment
type Comment struct {
	// The unique ID of the comment
	ID primitive.ObjectID `bson:"_id,omitempty" json:"id"`

	// The ID of the todo this comment belongs to
	TodoID primitive.ObjectID `bson:"todoId" json:"todoId"`

	// The ID of the user who authored the comment
	AuthorID primitive.ObjectID `bson:"authorId" json:"authorId"`

	// The body text of the comment
	Body string `bson:"body" json:"body"`

	// The date and time when the comment was created
	CreatedAt time.Time `bson:"createdAt" json:"createdAt"`
}
