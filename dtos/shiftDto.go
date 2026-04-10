package dtos

type CreateShiftReq struct {
	Title     string `json:"title" binding:"required"`
	StartTime string `json:"start_time" binding:"required"`
	EndTime   string `json:"end_time" binding:"required"`
	IsActive  *bool  `json:"is_active"`
}

type UpdateShiftReq struct {
	Title     string `json:"title"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
	IsActive  *bool  `json:"is_active"`
}

type ShiftRes struct {
	ID         string `json:"id"`
	Title      string `json:"title"`
	StartTime  string `json:"start_time"`
	EndTime    string `json:"end_time"`
	IsActive   bool   `json:"is_active"`
	BusinessID uint   `json:"business_id"`
	BranchID   uint   `json:"branch_id"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
}
