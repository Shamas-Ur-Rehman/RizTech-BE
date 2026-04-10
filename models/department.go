package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Department struct {
	ID                  primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	NameEn              string             `bson:"name_en" json:"name_en" validate:"required"`
	NameAr              string             `bson:"name_ar" json:"name_ar" validate:"required"`
	DescriptionEn       string             `bson:"description_en,omitempty" json:"description_en,omitempty"`
	DescriptionAr       string             `bson:"description_ar,omitempty" json:"description_ar,omitempty"`
	DepartmentShortName string             `bson:"department_short_name" json:"department_short_name"`
	ERPBranchID         uint               `bson:"erp_branch_id,omitempty" json:"erp_branch_id,omitempty"`
	IsCovered           bool               `bson:"is_covered,omitempty" json:"is_covered,omitempty"`
	BusinessID          uint               `bson:"business_id" json:"business_id"`
	BranchID            uint               `bson:"branch_id" json:"branch_id"`
	CreatedAt           time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt           time.Time          `bson:"updated_at" json:"updated_at"`
	DeletedAt           *time.Time         `bson:"deleted_at,omitempty" json:"deleted_at,omitempty"`
}
