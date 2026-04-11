package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Action values: get, create, update, delete
type Permission struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	RoleID    primitive.ObjectID `bson:"role_id" json:"role_id"`
	Module    string             `bson:"module" json:"module"`
	Action    string             `bson:"action" json:"action"` // get | create | update | delete
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
}
