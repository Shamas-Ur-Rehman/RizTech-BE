package models

import "time"

type Branch struct {
	ID         uint       `json:"id" gorm:"primaryKey"`
	NameEn     string     `json:"name_en" binding:"required"`
	NameAr     string     `json:"name_ar" binding:"required"`
	BusinessId uint       `json:"business_id" binding:"required"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	DeletedAt  *time.Time `json:"deleted_at,omitempty" gorm:"index"`
}
