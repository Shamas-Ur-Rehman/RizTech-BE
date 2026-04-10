package dtos

type CreateDepartmentReq struct {
	NameEn              string `json:"name_en" binding:"required"`
	NameAr              string `json:"name_ar" binding:"required"`
	DescriptionEn       string `json:"description_en"`
	DescriptionAr       string `json:"description_ar"`
	DepartmentShortName string `json:"department_short_name"`
	ERPBranchID         uint   `json:"erp_branch_id"`
	IsCovered           bool   `json:"is_covered"`
}

type UpdateDepartmentReq struct {
	NameEn              string `json:"name_en"`
	NameAr              string `json:"name_ar"`
	DescriptionEn       string `json:"description_en"`
	DescriptionAr       string `json:"description_ar"`
	DepartmentShortName string `json:"department_short_name"`
	ERPBranchID         uint   `json:"erp_branch_id"`
	IsCovered           *bool  `json:"is_covered"`
}

type DepartmentRes struct {
	ID                  string `json:"id"`
	NameEn              string `json:"name_en"`
	NameAr              string `json:"name_ar"`
	DescriptionEn       string `json:"description_en,omitempty"`
	DescriptionAr       string `json:"description_ar,omitempty"`
	DepartmentShortName string `json:"department_short_name"`
	ERPBranchID         uint   `json:"erp_branch_id,omitempty"`
	IsCovered           bool   `json:"is_covered"`
	BusinessID          uint   `json:"business_id"`
	BranchID            uint   `json:"branch_id"`
	CreatedAt           string `json:"created_at"`
	UpdatedAt           string `json:"updated_at"`
}
