package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Schedule struct {
	Day    string               `json:"day" bson:"day" validate:"required"`
	Shifts []primitive.ObjectID `json:"shifts" bson:"shifts"`
}

type Staff struct {
	ID             primitive.ObjectID   `bson:"_id,omitempty" json:"id,omitempty"`
	UserID         uint                 `bson:"user_id" json:"user_id" validate:"required"`
	BusinessID     uint                 `bson:"business_id" json:"business_id" validate:"required"`
	BranchID       uint                 `bson:"branch_id" json:"branch_id" validate:"required"`
	DepartmentID   primitive.ObjectID   `bson:"department_id,omitempty" json:"department_id,omitempty"`
	SpecialityID   primitive.ObjectID   `bson:"speciality_id,omitempty" json:"speciality_id,omitempty"`
	Schedule       []Schedule           `bson:"schedule,omitempty" json:"schedule,omitempty"`
	AssignedNurses []primitive.ObjectID `bson:"assigned_nurses,omitempty" json:"assigned_nurses,omitempty"`
	WardNo         string               `bson:"ward_no,omitempty" json:"ward_no,omitempty"`
	RoomNo         string               `bson:"room_no,omitempty" json:"room_no,omitempty"`
	Available      bool                 `bson:"available,omitempty" json:"available,omitempty"`
	CreatedAt      time.Time            `bson:"created_at" json:"created_at"`
	UpdatedAt      time.Time            `bson:"updated_at" json:"updated_at"`
	DeletedAt      *time.Time           `bson:"deleted_at,omitempty" json:"deleted_at,omitempty"`
}

type StaffResponse struct {
	ID             primitive.ObjectID   `bson:"_id,omitempty" json:"id,omitempty"`
	UserID         uint                 `bson:"user_id" json:"user_id"`
	User           *User                `json:"user,omitempty"`
	BusinessID     uint                 `bson:"business_id" json:"business_id"`
	BranchID       uint                 `bson:"branch_id" json:"branch_id"`
	DepartmentID   primitive.ObjectID   `bson:"department_id,omitempty" json:"department_id,omitempty"`
	Department     *Department          `json:"department,omitempty"`
	SpecialityID   primitive.ObjectID   `bson:"speciality_id,omitempty" json:"speciality_id,omitempty"`
	Speciality     *Speciality          `json:"speciality,omitempty"`
	Schedule       []Schedule           `bson:"schedule,omitempty" json:"schedule,omitempty"`
	AssignedNurses []primitive.ObjectID `bson:"assigned_nurses,omitempty" json:"assigned_nurses,omitempty"`
	Nurses         []*Staff             `json:"nurses,omitempty"`
	WardNo         string               `bson:"ward_no,omitempty" json:"ward_no,omitempty"`
	RoomNo         string               `bson:"room_no,omitempty" json:"room_no,omitempty"`
	Available      bool                 `bson:"available,omitempty" json:"available,omitempty"`
	CreatedAt      time.Time            `bson:"created_at" json:"created_at"`
	UpdatedAt      time.Time            `bson:"updated_at" json:"updated_at"`
	DeletedAt      *time.Time           `bson:"deleted_at,omitempty" json:"deleted_at,omitempty"`
}
