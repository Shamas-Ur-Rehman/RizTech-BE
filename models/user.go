package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Role name constants — used for seeding and checks
const (
	RoleAdmin = "admin"
	RoleUser  = "user"
)

type User struct {
	ID               primitive.ObjectID  `bson:"_id,omitempty"          json:"id"`
	FullName         string              `bson:"full_name"              json:"full_name"`
	Email            string              `bson:"email"                  json:"email"`
	Password         string              `bson:"password"               json:"-"`
	Contact          string              `bson:"contact,omitempty"      json:"contact,omitempty"`
	RoleID           primitive.ObjectID  `bson:"role_id"                json:"role_id"`
	RoleName         string              `bson:"role_name"              json:"role_name"`
	IsActive         bool                `bson:"is_active"              json:"is_active"`
	ResetToken       string              `bson:"reset_token,omitempty"  json:"-"`
	ResetTokenExpiry *time.Time          `bson:"reset_token_expiry,omitempty" json:"-"`
	CreatedAt        time.Time           `bson:"created_at"             json:"created_at"`
	UpdatedAt        time.Time           `bson:"updated_at"             json:"updated_at"`
}
