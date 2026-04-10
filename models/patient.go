package models

import (
	"time"
)

type Patient struct {
	PatientID               string             `bson:"patient_id" json:"patient_id"`
	OutPatientID            string             `bson:"out_patient_id" json:"out_patient_id,omitempty"`
	FullName                string             `bson:"full_name" json:"full_name" validate:"required"`
	Contact                 string             `bson:"contact" json:"contact" validate:"required"`
	ContactNumber           string             `bson:"contact_number" json:"contact_number,omitempty"`
	DocumentID              string             `bson:"document_id" json:"document_id" validate:"required"`
	RCMRef                  string             `bson:"rcm_ref" json:"rcm_ref,omitempty"`
	ERPRef                  string             `bson:"erp_ref" json:"erp_ref,omitempty"`
	DateType                string             `bson:"date_type" json:"date_type"`
	UserID                  uint               `bson:"user_id,omitempty" json:"user_id,omitempty"`
	Maternity               string             `bson:"maternity,omitempty" json:"maternity,omitempty"`
	MartialStatus           string             `bson:"martial_status,omitempty" json:"martial_status,omitempty"`
	Gender                  string             `bson:"gender,omitempty" json:"gender,omitempty"`
	VisaNumber              string             `bson:"visa_no,omitempty" json:"visa_no,omitempty"`
	FileNumber              string             `bson:"file_no" json:"file_no" validate:"required"`
	MRN                     int64              `bson:"mrn,omitempty" json:"mrn,omitempty"`
	PatientType             string             `bson:"beneficiary_type,omitempty" json:"beneficiary_type,omitempty"`
	BusinessID              uint               `bson:"business_id" json:"business_id" validate:"required"`
	BranchID                uint               `bson:"branch_id" json:"branch_id" validate:"required"`
	City                    string             `bson:"city,omitempty" json:"city,omitempty"`
	ResidencyType           string             `bson:"residency_type,omitempty" json:"residency_type,omitempty"`
	InsurancePlan           []InsurancePlan    `bson:"insurance_plans,omitempty" json:"insurance_plans,omitempty"`
	Guarantors              []Guarantor        `bson:"guarantors,omitempty" json:"guarantors,omitempty"`
	EmergencyContacts       []EmergencyContact `bson:"emergency_contacts,omitempty" json:"emergency_contacts,omitempty"`
	Address                 string             `bson:"address,omitempty" json:"address,omitempty"`
	VisaType                string             `bson:"visa_type,omitempty" json:"visa_type,omitempty"`
	VisaTitle               string             `bson:"visa_title,omitempty" json:"visa_title,omitempty"`
	DocumentType            string             `bson:"document_type,omitempty" json:"document_type,omitempty"`
	PassportNumber          string             `bson:"passport_no,omitempty" json:"passport_no,omitempty"`
	InsuranceDuration       string             `bson:"insurance_duration,omitempty" json:"insurance_duration,omitempty"`
	BorderNumber            string             `bson:"border_no,omitempty" json:"border_no,omitempty"`
	BirthDate               string             `bson:"dob,omitempty" json:"dob,omitempty"`
	Nationality             string             `bson:"nationality,omitempty" json:"nationality,omitempty"`
	Profession              string             `bson:"profession,omitempty" json:"profession,omitempty"`
	Religion                any                `bson:"religion,omitempty" json:"religion,omitempty"`
	BloodGroup              string             `bson:"blood_group,omitempty" json:"blood_group,omitempty"`
	PreferredLanguage       string             `bson:"preferred_language,omitempty" json:"preferred_language,omitempty"`
	EmergencyNumber         string             `bson:"emergency_number,omitempty" json:"emergency_number,omitempty"`
	EHealthID               string             `bson:"eHealth_id,omitempty" json:"eHealth_id,omitempty"`
	Occupation              string             `bson:"occupation,omitempty" json:"occupation,omitempty"`
	SubscriberID            string             `bson:"subscriber_id,omitempty" json:"subscriber_id,omitempty"`
	SubscriberRelationship  string             `bson:"subscriber_relationship,omitempty" json:"subscriber_relationship,omitempty"`
	SubscriberInsurancePlan []string           `bson:"subscriber_insurance_plan,omitempty" json:"subscriber_insurance_plan,omitempty"`
	IsNewBorn               bool               `bson:"is_new_born,omitempty" json:"is_new_born,omitempty"`
	Password                string             `bson:"password,omitempty" json:"password,omitempty"`
	Email                   string             `bson:"email,omitempty" json:"email,omitempty" validate:"omitempty,email"`
	UserLogin               bool               `bson:"user_login,omitempty" json:"user_login,omitempty"`
	Role                    string             `bson:"role,omitempty" json:"role,omitempty"`
	SubscriberIdAlt         string             `bson:"subscriberId,omitempty" json:"subscriberId,omitempty"`
	SubscriberRelAlt        string             `bson:"subscriberRelationship,omitempty" json:"subscriberRelationship,omitempty"`
	SubscriberPlanAlt       []string           `bson:"subscriberInsurancePlan,omitempty" json:"subscriberInsurancePlan,omitempty"`
	CreatedAt               time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt               time.Time          `bson:"updated_at" json:"updated_at"`
	DeletedAt               *time.Time         `bson:"deleted_at,omitempty" json:"deleted_at,omitempty"`
	PragnancyHistory        []PragnancyHistory `bson:"pragnancy_history,omitempty" json:"pragnancy_history,omitempty"`
}

func (p *Patient) Age() int {
	if p.BirthDate == "" {
		return -1
	}
	birthTime, err := time.Parse("2006-01-02", p.BirthDate)
	if err != nil {
		return -1
	}
	now := time.Now()
	age := now.Year() - birthTime.Year()
	if now.Month() < birthTime.Month() || (now.Month() == birthTime.Month() && now.Day() < birthTime.Day()) {
		age--
	}
	return age
}

type InsurancePlan struct {
	ID                     string           `json:"id,omitempty"`
	InsurancePlanID        string           `bson:"insurance_plan_id" json:"insurance_plan_id"`
	MemberCardId           string           `bson:"member_card_id" json:"member_card_id"`
	PolicyNumber           string           `bson:"policy_number" json:"policy_number"`
	ExpiryDate             string           `bson:"expiry_date" json:"expiry_date"`
	IsPrimary              bool             `bson:"is_primary" json:"is_primary"`
	PayerId                string           `bson:"payer_id" json:"payer_id"`
	HisPayerId             string           `bson:"his_payer_id" json:"his_payer_id"`
	PayerName              string           `bson:"payer_name" json:"payer_name"`
	RelationWithSubscriber string           `bson:"relation_with_subscriber" json:"relation_with_subscriber"`
	CoverageType           string           `bson:"coverage_type" json:"coverage_type"`
	PatientShare           string           `bson:"patient_share" json:"patient_share"`
	MaxLimit               int              `bson:"max_limit" json:"max_limit"`
	DiscountPercentage     *int             `bson:"discount_percentage,omitempty" json:"discount_percentage,omitempty"`
	RemainingLimit         float64          `bson:"remaining_limit" json:"remaining_limit"`
	Network                string           `bson:"network" json:"network"`
	IssueDate              string           `bson:"issue_date" json:"issue_date"`
	SponsorNo              string           `bson:"sponsor_no" json:"sponsor_no"`
	PolicyClass            string           `bson:"policy_class" json:"policy_class"`
	PolicyHolder           string           `bson:"policy_holder" json:"policy_holder"`
	InsuranceType          string           `bson:"insurance_type" json:"insurance_type"`
	InsuranceStatus        string           `bson:"insurance_status" json:"insurance_status"`
	InsuranceDuration      string           `bson:"insurance_duration" json:"insurance_duration"`
	ClassID                string           `bson:"class_id" json:"class_id"`
	ClassName              string           `bson:"class_name" json:"class_name"`
	ClassType              string           `bson:"class_type" json:"class_type"`
	PolicyClassType        string           `json:"policy_class_type" bson:"policy_class_type"`
	PolicyClassID          string           `json:"policy_class_id" bson:"policy_class_id"`
	PolicyClassName        string           `json:"policy_class_name" bson:"policy_class_name"`
	Payer                  *Payer           `bson:"payer" json:"payer"`
	PolicyObject           *InsurancePolicy `json:"policy_object,omitempty"`
	EligiblityStatus       string           `json:"eligiblity_status,omitempty" bson:"eligiblity_status,omitempty"`
	Disposition            string           `json:"disposition,omitempty" bson:"disposition,omitempty"`
	Period                 Period           `json:"period,omitempty" bson:"period,omitempty"`
}

type Guarantor struct {
	RelationWithPatient string `json:"relation_with_patient,omitempty" bson:"relation_with_patient,omitempty"`
	Relationship        string `json:"relationship,omitempty" bson:"relationship,omitempty"`
	Name                string `json:"name" bson:"name"`
	Gender              string `json:"gender,omitempty" bson:"gender,omitempty"`
	Phone               string `json:"phone,omitempty" bson:"phone,omitempty"`
	ContactNumber       string `json:"contact_number,omitempty" bson:"contact_number,omitempty"`
	DOB                 string `json:"dob,omitempty" bson:"dob,omitempty"`
	Address             string `json:"address,omitempty" bson:"address,omitempty"`
	DocumentID          string `json:"document_id,omitempty" bson:"document_id,omitempty"`
	DocumentType        string `json:"document_type,omitempty" bson:"document_type,omitempty"`
	Nationality         string `json:"nationality,omitempty" bson:"nationality,omitempty"`
	City                string `json:"city,omitempty" bson:"city,omitempty"`
}

type EmergencyContact struct {
	Name          string `json:"name,omitempty" bson:"name,omitempty"`
	FirstName     string `json:"first_name,omitempty" bson:"first_name,omitempty"`
	LastName      string `json:"last_name,omitempty" bson:"last_name,omitempty"`
	PhoneNumber   string `json:"phone_number,omitempty" bson:"phone_number,omitempty"`
	ContactNumber string `json:"contact_number,omitempty" bson:"contact_number,omitempty"`
	Email         string `json:"email,omitempty" bson:"email,omitempty"`
	Address       string `json:"address,omitempty" bson:"address,omitempty"`
	Relationship  string `json:"relationship,omitempty" bson:"relationship,omitempty"`
	Type          string `json:"type,omitempty" bson:"type,omitempty"`
	Comments      string `json:"comments,omitempty" bson:"comments,omitempty"`
}

type PragnancyHistory struct {
	LMP string `bson:"lmp" json:"lmp"`
	EDD string `bson:"edd" json:"edd"`
}

type InsurancePolicy struct {
	ID               string                     `bson:"_id,omitempty" json:"id"`
	Payer            string                     `bson:"payer" json:"payer"`
	PharmacyDetails  PharmacyPoliciesDeductable `bson:"pharmacy_details" json:"pharmacy_details"`
	InsuranceType    string                     `bson:"insuranceType" json:"insurance_type"`
	PolicyNumber     string                     `bson:"policyNumber" json:"policy_number"`
	PolicyClass      string                     `bson:"policyClass" json:"policy_class"`
	PolicyHolderName string                     `bson:"policyHolderName" json:"policy_holder_name"`
	PolicyNetwork    string                     `bson:"policyNetwork" json:"policy_network"`
	StartDate        string                     `bson:"startDate" json:"start_date"`
	EndDate          string                     `bson:"endDate" json:"end_date"`
}

type Period struct {
	Start string `json:"start"`
	End   string `json:"end"`
}
