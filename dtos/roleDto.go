package dtos

import "time"

type RoleReq struct {
	Name       string `json:"name" binding:"required"`
	BusinessId uint   `json:"business_id"`
	BranchId   uint   `json:"branch_id"`
}

type RoleRes struct {
	ID          uint             `json:"id"`
	Name        string           `json:"name"`
	BusinessId  uint             `json:"business_id"`
	BranchId    uint             `json:"branch_id"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
	Permissions []PermissionRes  `json:"permissions,omitempty"`
}

type RoleUpdateReq struct {
	Name       string `json:"name"`
	BusinessId uint   `json:"business_id"`
	BranchId   uint   `json:"branch_id"`
}

type AssignPermissionsReq struct {
	RoleID      uint                    `json:"role_id" binding:"required"`
	Permissions []ModulePermissionsReq  `json:"permissions" binding:"required"`
}

type ModulePermissionsReq struct {
	ModuleID uint     `json:"module_id" binding:"required"`
	Actions  []string `json:"actions" binding:"required"` // ["get", "create", "update", "delete"]
}

type RemovePermissionsReq struct {
	RoleID        uint   `json:"role_id" binding:"required"`
	PermissionIDs []uint `json:"permission_ids" binding:"required"`
}
