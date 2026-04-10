package dtos

import "time"

type PermissionReq struct {
	RoleID     uint   `json:"role_id" binding:"required"`
	ModuleID   uint   `json:"module_id" binding:"required"`
	Action     string `json:"action" binding:"required"`
	BusinessId uint   `json:"business_id"`
	BranchId   uint   `json:"branch_id"`
}

type PermissionRes struct {
	ID         uint      `json:"id"`
	RoleID     uint      `json:"role_id"`
	RoleName   string    `json:"role_name,omitempty"`
	ModuleID   uint      `json:"module_id"`
	ModuleName string    `json:"module_name,omitempty"`
	Action     string    `json:"action"`
	BusinessId uint      `json:"business_id"`
	BranchId   uint      `json:"branch_id"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
