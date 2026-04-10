package models

import "time"

type Permission struct {
	ID         uint       `json:"id" gorm:"primaryKey;autoIncrement"`
	RoleID     uint       `json:"role_id" gorm:"not null;index:idx_role_module_action,priority:1"`
	ModuleID   uint       `json:"module_id" gorm:"not null;index:idx_role_module_action,priority:2"`
	Action     string     `json:"action" gorm:"not null;type:ENUM('get','create','update','delete');index:idx_role_module_action,priority:3"`
	BusinessId uint       `json:"business_id"`
	BranchId   uint       `json:"branch_id"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	DeletedAt  *time.Time `json:"deleted_at,omitempty" gorm:"index"`
	Role       Role       `gorm:"foreignKey:RoleID;references:ID"`
	Module     Module     `gorm:"foreignKey:ModuleID;references:ID"`
}

func (Permission) TableName() string {
	return "inpatient_permissions"
}
