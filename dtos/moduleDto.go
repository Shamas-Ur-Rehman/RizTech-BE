package dtos

import "time"

type ModuleReq struct {
	Name        string `json:"name" binding:"required"`
	DisplayName string `json:"display_name" binding:"required"`
	Description string `json:"description"`
	BusinessId  uint   `json:"business_id"`
	BranchId    uint   `json:"branch_id"`
}

type ModuleRes struct {
	ID          uint      `json:"id"`
	Name        string    `json:"name"`
	DisplayName string    `json:"display_name"`
	Description string    `json:"description"`
	BusinessId  uint      `json:"business_id"`
	BranchId    uint      `json:"branch_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ModuleUpdateReq struct {
	DisplayName string `json:"display_name"`
	Description string `json:"description"`
	BusinessId  uint   `json:"business_id"`
	BranchId    uint   `json:"branch_id"`
}
