package dtos

type CreateSpecialityReq struct {
	NameEn       string `json:"name_en" binding:"required"`
	NameAr       string `json:"name_ar" binding:"required"`
	Description  string `json:"description"`
	DepartmentID string `json:"department_id" binding:"required"`
}

type UpdateSpecialityReq struct {
	NameEn       string `json:"name_en"`
	NameAr       string `json:"name_ar"`
	Description  string `json:"description"`
	DepartmentID string `json:"department_id"`
}

type SpecialityRes struct {
	ID             string              `json:"id"`
	NameEn         string              `json:"name_en"`
	NameAr         string              `json:"name_ar"`
	Description    string              `json:"description,omitempty"`
	DepartmentID   string              `json:"department_id"`
	Department     *DepartmentBasicRes `json:"department,omitempty"`
	BusinessID     uint                `json:"business_id"`
	BranchID       uint                `json:"branch_id"`
	CreatedAt      string              `json:"created_at"`
	UpdatedAt      string              `json:"updated_at"`
}

type DepartmentBasicRes struct {
	ID                  string `json:"id"`
	NameEn              string `json:"name_en"`
	NameAr              string `json:"name_ar"`
	DepartmentShortName string `json:"department_short_name,omitempty"`
}
