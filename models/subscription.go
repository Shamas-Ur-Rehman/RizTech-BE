package models

import "time"

type Subscription struct {
	ID              uint       `json:"id" gorm:"primaryKey"`
	BusinessId      uint       `json:"business_id" binding:"required"`
	ApiKey          string     `json:"api_key"`
	SecretKey       string     `json:"secret_key"`
	RadiologyUrl    string     `json:"radiology_url"`
	RadiologyApiKey string     `json:"radiology_api_key"`
	LabUrl          string     `json:"lab_url"`
	LabApiKey       string     `json:"lab_api_key"`
	HisUrl          string     `json:"his_url"`
	HisApiKey       string     `json:"his_api_key"`
	BranchId        uint       `json:"branch_id"`
	ExpiryDate      time.Time  `json:"expiry_date"`
	Status          string     `json:"status"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	DeletedAt       *time.Time `json:"deleted_at,omitempty" gorm:"index"`
}

func (Subscription) TableName() string {
	return "inpatient_subscriptions"
}
