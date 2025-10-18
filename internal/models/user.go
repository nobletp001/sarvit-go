package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User represents an application user.
//
// swagger:model User
type User struct {
	// The unique ID of the user
	ID primitive.ObjectID `bson:"_id,omitempty" json:"id"`

	// The full name of the user
	Name string `bson:"name" json:"name"`

	// The user's email address (used for login)
	Email string `bson:"email" json:"email"`

	// The user's password (never exposed in JSON responses)
	Password string `bson:"password" json:"-"`

	// The date and time when the user account was created
	CreatedAt time.Time `bson:"createdAt" json:"createdAt"`
}
