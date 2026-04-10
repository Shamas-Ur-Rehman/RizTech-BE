package dtos

type UserReq struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type UserRes struct {
	ID          uint    `json:"id"`
	Email       string  `json:"email"`
	FullName    string  `json:"full_name"`
	EmployeeId  string  `json:"employee_id"`
	Contact     string  `json:"contact"`
	DateOfBirth *string `json:"date_of_birth,omitempty"`
	Age         *int    `json:"age,omitempty"`
	Gender      string  `json:"gender,omitempty"`
	DocumentId  string  `json:"document_id"`
	License     string  `json:"license,omitempty"`
	Address     string  `json:"address"`
	Nationality string  `json:"nationality"`
	RoleID      uint    `json:"role_id"`
	RoleName    string  `json:"role_name"`
	Service     string  `json:"service"`
	BranchId    uint    `json:"branch_id"`
	BusinessId  uint    `json:"business_id"`
	FCMToken    string  `json:"fcm_token"`
	IsActive    bool    `json:"is_active"`
	IsStaff     bool    `json:"is_staff"`
}

type ChangePasswordReq struct {
	UserID      uint   `json:"user_id" binding:"required"`
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}
