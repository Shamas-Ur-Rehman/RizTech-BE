package models

import (
	"time"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Shift struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Title      string             `bson:"title" json:"title" validate:"required"`
	StartTime  string             `bson:"start_time" json:"start_time" validate:"required"`
	EndTime    string             `bson:"end_time" json:"end_time" validate:"required"`
	IsActive   bool               `bson:"is_active" json:"is_active"`
	BusinessID uint               `bson:"business_id" json:"business_id"`
	BranchID   uint               `bson:"branch_id" json:"branch_id"`
	CreatedAt  time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt  time.Time          `bson:"updated_at" json:"updated_at"`
	DeletedAt  *time.Time         `bson:"deleted_at,omitempty" json:"deleted_at,omitempty"`
}
