package models

import "time"

type Role struct {
	ID          uint         `json:"id" gorm:"primaryKey"`
	Name        string       `json:"name" binding:"required" gorm:"unique"`
	BusinessId  uint         `json:"business_id"`
	BranchId    uint         `json:"branch_id"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
	DeletedAt   *time.Time   `json:"deleted_at,omitempty" gorm:"index"`
	Permissions []Permission `gorm:"foreignKey:RoleID"`
}

func (Role) TableName() string {
	return "inpatient_roles"
}
