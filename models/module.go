package models

import "time"

type Module struct {
	ID          uint         `json:"id" gorm:"primaryKey"`
	Name        string       `json:"name" gorm:"unique;not null"`
	DisplayName string       `json:"display_name" gorm:"not null"`
	Description string       `json:"description"`
	BusinessId  uint         `json:"business_id"`
	BranchId    uint         `json:"branch_id"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
	DeletedAt   *time.Time   `json:"deleted_at,omitempty" gorm:"index"`
	Permissions []Permission `gorm:"foreignKey:ModuleID"`
}

func (Module) TableName() string {
	return "inpatient_modules"
}
