package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Todo represents a post created by a user.
//
// swagger:model Todo
type Todo struct {
	// The unique ID of the todo
	ID primitive.ObjectID `bson:"_id,omitempty" json:"id"`

	// The ID of the user who created the todo
	AuthorID primitive.ObjectID `bson:"authorId" json:"authorId"`

	// The title of the todo
	Title string `bson:"title" json:"title"`

	// The body or content of the todo
	Body string `bson:"body" json:"body"`

	// A list of user IDs who liked this todo
	Likes []primitive.ObjectID `bson:"likes" json:"likes"`

	// The date and time when the todo was created
	CreatedAt time.Time `bson:"createdAt" json:"createdAt"`

	// The date and time when the todo was last updated
	UpdatedAt time.Time `bson:"updatedAt" json:"updatedAt"`
}
