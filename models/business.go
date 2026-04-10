package models

import "time"

type Business struct {
	ID                  uint           `json:"id" gorm:"primaryKey;autoIncrement"`
	NameEn              string         `json:"name_en" gorm:"unique" validate:"required"`
	NameAr              string         `json:"name_ar" gorm:"unique" validate:"required"`
	BusinessType        string         `gorm:"type:enum('clinic','polyclinic','hospital');not null" json:"business_type"`
	Address             string         `json:"address"`
	ContactInfo         string         `json:"contact_info"`
	VATNo               string         `json:"vat_no" gorm:"unique" validate:"required"`
	CRNo                string         `json:"cr_no" gorm:"unique" validate:"required"`
	City                string         `json:"city"`
	VATRegistrationDate string         `json:"vat_registration_date" validate:"required"`
	Logo                string         `json:"logo" validate:"required"`
	Signature           string         `json:"signature"`
	Stamp               string         `json:"stamp"`
	RCMEmail            string         `json:"rcm_email" validate:"required"`
	RCMPassword         string         `json:"rcm_password" validate:"required"`
	Latitude            string         `json:"latitude"`
	Longitude           string         `json:"longitude"`
	CreatedAt           time.Time      `bson:"createdAt" json:"created_at"`
	UpdatedAt           time.Time      `bson:"updatedAt" json:"updated_at"`
	LicenseNo           string         `bson:"license_no" json:"license_no"`
	Email               string         `bson:"email" json:"email"`
	RegistrationNo      string         `json:"registration_no"`
	PrivacyPolicy       string         `json:"privacy_policy" gorm:"column:privacy_policy;type:text"`
	InPatientDB         string         `json:"in_patient_db" gorm:"unique;not null"`
	Subscription        *Subscription  `json:"subscription,omitempty" gorm:"foreignKey:BusinessId"`
}
