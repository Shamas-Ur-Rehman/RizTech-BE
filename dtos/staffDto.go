package dtos

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"supergit/inpatient/models"
)

type ScheduleReq struct {
	Day    string   `json:"day" validate:"required"`
	Shifts []string `json:"shifts"`
}

type CreateUserData struct {
	FullName    string  `json:"full_name" validate:"required"`
	Email       string  `json:"email" validate:"required,email"`
	Password    string  `json:"password" validate:"required,min=8"`
	Contact     string  `json:"contact"`
	DateOfBirth *string `json:"date_of_birth"`
	Gender      string  `json:"gender"`
	RoleID      uint    `json:"role_id" validate:"required"`
	Address     string  `json:"address"`
	Nationality string  `json:"nationality"`
	DocumentId  string  `json:"document_id"`
	License     string  `json:"license"`
}

type CreateStaffReq struct {
	UserData *CreateUserData `json:"user_data" validate:"required"`
	DepartmentID string `json:"department_id,omitempty"`
	SpecialityID string `json:"speciality_id,omitempty"`
	Schedule []ScheduleReq `json:"schedule,omitempty"`
	AssignedNurses []string `json:"assigned_nurses,omitempty"`
	WardNo string `json:"ward_no,omitempty"`
	RoomNo string `json:"room_no,omitempty"`
	Available bool `json:"available,omitempty"`
}

type UpdateStaffReq struct {
	DepartmentID   string        `json:"department_id,omitempty"`
	SpecialityID   string        `json:"speciality_id,omitempty"`
	Schedule       []ScheduleReq `json:"schedule,omitempty"`
	AssignedNurses []string      `json:"assigned_nurses,omitempty"`
	WardNo         string        `json:"ward_no,omitempty"`
	RoomNo         string        `json:"room_no,omitempty"`
	Available      *bool         `json:"available,omitempty"`
}

type StaffRes struct {
	ID           primitive.ObjectID   `json:"id,omitempty"`
	UserID       uint                 `json:"user_id"`
	User         *UserRes             `json:"user,omitempty"`
	BusinessID   uint                 `json:"business_id"`
	BranchID     uint                 `json:"branch_id"`
	
	DepartmentID primitive.ObjectID   `json:"department_id,omitempty"`
	Department   *DepartmentRes       `json:"department,omitempty"`
	SpecialityID primitive.ObjectID   `json:"speciality_id,omitempty"`
	Speciality   *SpecialityRes       `json:"speciality,omitempty"`
	
	Schedule       []models.Schedule    `json:"schedule,omitempty"`
	AssignedNurses []primitive.ObjectID `json:"assigned_nurses,omitempty"`
	Nurses         []*StaffRes          `json:"nurses,omitempty"`
	WardNo         string               `json:"ward_no,omitempty"`
	RoomNo         string               `json:"room_no,omitempty"`
	Available      bool                 `json:"available,omitempty"`
	
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}
