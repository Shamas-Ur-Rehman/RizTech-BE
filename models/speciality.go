package models

import (
	"time"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Speciality struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	NameEn       string             `bson:"name_en" json:"name_en" validate:"required"`
	NameAr       string             `bson:"name_ar" json:"name_ar" validate:"required"`
	Description  string             `bson:"description,omitempty" json:"description,omitempty"`
	DepartmentID primitive.ObjectID `bson:"department_id" json:"department_id" validate:"required"`
	BusinessID   uint               `bson:"business_id" json:"business_id"`
	BranchID     uint               `bson:"branch_id" json:"branch_id"`
	CreatedAt    time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt    time.Time          `bson:"updated_at" json:"updated_at"`
	DeletedAt    *time.Time         `bson:"deleted_at,omitempty" json:"deleted_at,omitempty"`
}
