package models

import "time"

type User struct {
	ID          uint       `json:"id" gorm:"primaryKey"`
	FullName    string     `json:"full_name" binding:"required"`
	Email       string     `json:"email" binding:"required,email" gorm:"unique"`
	Password    string     `json:"password" binding:"required,min=8"`
	EmployeeId  string     `json:"employee_id"`
	Contact     string     `json:"contact"`
	DateOfBirth *time.Time `json:"date_of_birth"`
	Gender      string     `json:"gender" gorm:"type:ENUM('male','female','other')"`
	RoleID      uint       `json:"role_id" binding:"required"`
	IsActive    bool       `json:"is_active" gorm:"default:true"`
	IsStaff     bool       `json:"is_staff" gorm:"default:false"`
	Service     string     `json:"service" gorm:"type:ENUM('inpatient','radiology');default:'inpatient'"`
	Address     string     `json:"address"`
	Nationality string     `json:"nationality"`
	DocumentId  string     `json:"document_id"`
	License     string     `json:"license"`
	FCMToken    string     `json:"fcm_token"`
	DeviceId    string     `json:"device_id"`
	DeviceType  string     `json:"device_type"`
	BusinessId  uint       `json:"business_id"`
	BranchId    uint       `json:"branch_id"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty" gorm:"index"`

	Business Business `gorm:"foreignKey:BusinessId;references:ID"`
	Role     Role     `gorm:"foreignKey:RoleID;references:ID"`
}

func (User) TableName() string {
	return "inpatient_users"
}
