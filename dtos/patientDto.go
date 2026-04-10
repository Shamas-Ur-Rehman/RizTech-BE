package dtos

import "supergit/inpatient/models"

type CreatePatientDto struct {
	FullName                string                    `json:"full_name" validate:"required,min=3"`
	Contact                 string                    `json:"contact" validate:"required,min=10"`
	ContactNumber           string                    `json:"contact_number,omitempty"`
	DocumentID              string                    `json:"document_id" validate:"required,min=10"`
	Gender                  string                    `json:"gender,omitempty" validate:"omitempty,oneof=male female other unknown"`
	BirthDate               string                    `json:"dob,omitempty"`
	DateType                string                    `json:"date_type,omitempty"`
	Email                   string                    `json:"email,omitempty" validate:"omitempty,email"`
	FileNumber              string                    `json:"file_no" validate:"omitempty,min=3"`
	OutPatientID            string                    `json:"out_patient_id,omitempty"`
	RCMRef                  string                    `json:"rcm_ref,omitempty"`
	PatientType             string                    `json:"beneficiary_type,omitempty"`
	Nationality             string                    `json:"nationality,omitempty"`
	Address                 string                    `json:"address,omitempty"`
	City                    string                    `json:"city,omitempty"`
	BloodGroup              string                    `json:"blood_group,omitempty"`
	MartialStatus           string                    `json:"martial_status,omitempty" validate:"omitempty,oneof=D L M U W"`
	Occupation              string                    `json:"occupation,omitempty"`
	Religion                string                    `json:"religion,omitempty"`
	DocumentType            string                    `json:"document_type,omitempty"`
	ResidencyType           string                    `json:"residency_type,omitempty"`
	PassportNo              string                    `json:"passport_no,omitempty"`
	VisaTitle               string                    `json:"visa_title,omitempty"`
	VisaNo                  string                    `json:"visa_no,omitempty"`
	VisaType                string                    `json:"visa_type,omitempty"`
	BorderNo                string                    `json:"border_no,omitempty"`
	InsuranceDuration       string                    `json:"insurance_duration,omitempty"`
	IsNewBorn               bool                      `json:"is_new_born,omitempty"`
	Role                    string                    `json:"role,omitempty"`
	SubscriberID            string                    `json:"subscriber_id,omitempty"`
	SubscriberRelationship  string                    `json:"subscriber_relationship,omitempty"`
	SubscriberInsurancePlan []string                  `json:"subscriber_insurance_plan,omitempty"`
	InsurancePlan           []InsurancePlanDto        `json:"insurance_plans,omitempty"`
	Guarantors              []models.Guarantor        `json:"guarantors,omitempty"`
	EmergencyContacts       []models.EmergencyContact `json:"emergency_contacts,omitempty"`
	PragnancyHistory        []models.PragnancyHistory `json:"pragnancy_history,omitempty"`
}

type UpdatePatientDto struct {
	FullName                   string                    `json:"full_name,omitempty"`
	Contact                    string                    `json:"contact,omitempty"`
	ContactNumber              string                    `json:"contact_number,omitempty"`
	DocumentID                 string                    `json:"document_id,omitempty"`
	Gender                     string                    `json:"gender,omitempty" validate:"omitempty,oneof=male female other"`
	BirthDate                  string                    `json:"dob,omitempty"`
	DateType                   string                    `json:"date_type,omitempty"`
	Email                      string                    `json:"email,omitempty" validate:"omitempty,email"`
	FileNumber                 string                    `json:"file_no,omitempty"`
	PatientID                  string                    `json:"patient_id,omitempty"`
	RCMRef                     string                    `json:"rcm_ref,omitempty"`
	PatientType                string                    `json:"beneficiary_type,omitempty"`
	Nationality                string                    `json:"nationality,omitempty"`
	Address                    string                    `json:"address,omitempty"`
	City                       string                    `json:"city,omitempty"`
	BloodGroup                 string                    `json:"blood_group,omitempty"`
	MartialStatus              string                    `json:"martial_status,omitempty"`
	Occupation                 string                    `json:"occupation,omitempty"`
	Religion                   string                    `json:"religion,omitempty"`
	DocumentType               string                    `json:"document_type,omitempty"`
	ResidencyType              string                    `json:"residency_type,omitempty"`
	PassportNo                 string                    `json:"passport_no,omitempty"`
	VisaTitle                  string                    `json:"visa_title,omitempty"`
	VisaNo                     string                    `json:"visa_no,omitempty"`
	VisaType                   string                    `json:"visa_type,omitempty"`
	BorderNo                   string                    `json:"border_no,omitempty"`
	InsuranceDuration          string                    `json:"insurance_duration,omitempty"`
	SubscriberID               string                    `json:"subscriber_id,omitempty"`
	SubscriberRelationship     string                    `json:"subscriber_relationship,omitempty"`
	SubscriberInsurancePlan    []string                  `json:"subscriber_insurance_plan,omitempty"`
	SubscriberIdAlt            string                    `json:"subscriberId,omitempty"`
	SubscriberRelationshipAlt  string                    `json:"subscriberRelationship,omitempty"`
	SubscriberInsurancePlanAlt []string                  `json:"subscriberInsurancePlan,omitempty"`
	InsurancePlan              []InsurancePlanDto        `json:"insurance_plans,omitempty"`
	Guarantors                 []models.Guarantor        `json:"guarantors,omitempty"`
	EmergencyContacts          []models.EmergencyContact `json:"emergency_contacts,omitempty"`
	PragnancyHistory           []models.PragnancyHistory `json:"pragnancy_history,omitempty"`
}

type PatientResponseDto struct {
	PatientID         string                    `json:"patient_id"`
	OutPatientID      string                    `json:"out_patient_id"`
	FullName          string                    `json:"full_name"`
	Gender            string                    `json:"gender"`
	DocumentID        string                    `json:"document_id"`
	PatientType       string                    `json:"beneficiary_type"`
	FileNumber        string                    `json:"file_no"`
	Nationality       string                    `json:"nationality"`
	Address           string                    `json:"address"`
	ERPRef            string                    `json:"erp_ref"`
	BirthDate         string                    `json:"dob"`
	Contact           string                    `json:"contact"`
	Email             string                    `json:"email"`
	Age               int                       `json:"age"`
	BusinessID        uint                      `json:"business_id"`
	BranchID          uint                      `json:"branch_id"`
	InsurancePlan     []InsurancePlanResponse   `json:"insurance_plans,omitempty"`
	Guarantors        []models.Guarantor        `json:"guarantors,omitempty"`
	EmergencyContacts []models.EmergencyContact `json:"emergency_contacts,omitempty"`
	CreatedAt         string                    `json:"created_at"`
	UpdatedAt         string                    `json:"updated_at"`
}

type InsurancePlanDto struct {
	InsurancePlanID        string                  `json:"insurance_plan_id"`
	MemberCardId           string                  `json:"member_card_id,omitempty"`
	PolicyNumber           string                  `json:"policy_number"`
	ExpiryDate             string                  `json:"expiry_date"`
	IsPrimary              bool                    `json:"is_primary"`
	PayerId                string                  `json:"payer_id"`
	HisPayerId             string                  `bson:"his_payer_id" json:"his_payer_id"`
	PayerName              string                  `json:"payer_name,omitempty"`
	RelationWithSubscriber string                  `json:"relation_with_subscriber,omitempty"`
	CoverageType           string                  `json:"coverage_type,omitempty"`
	PatientShare           string                  `json:"patient_share,omitempty"`
	MaxLimit               int                     `json:"max_limit,omitempty"`
	DiscountPercentage     *int                    `json:"discount_percentage,omitempty"`
	RemainingLimit         float64                 `json:"remaining_limit,omitempty"`
	Network                string                  `json:"network,omitempty"`
	IssueDate              string                  `json:"issue_date,omitempty"`
	SponsorNo              string                  `bson:"sponsor_no" json:"sponsor_no"`
	PolicyClass            string                  `bson:"policy_class" json:"policy_class"`
	PolicyHolder           string                  `bson:"policy_holder" json:"policy_holder"`
	InsuranceType          string                  `bson:"insurance_type" json:"insurance_type"`
	InsuranceStatus        string                  `bson:"insurance_status" json:"insurance_status"`
	InsuranceDuration      string                  `bson:"insurance_duration" json:"insurance_duration"`
	ClassID                string                  `bson:"class_id" json:"class_id"`
	ClassName              string                  `bson:"class_name" json:"class_name"`
	ClassType              string                  `bson:"class_type" json:"class_type"`
	PolicyClassType        string                  `json:"policy_class_type" bson:"policy_class_type"`
	PolicyClassID          string                  `json:"policy_class_id" bson:"policy_class_id"`
	PolicyClassName        string                  `json:"policy_class_name" bson:"policy_class_name"`
	Payer                  *models.Payer           `bson:"payer" json:"payer"`
	PolicyObject           *models.InsurancePolicy `json:"policy_object,omitempty"`
	Period                 models.Period           `json:"period,omitempty" bson:"period,omitempty"`
}

type InsurancePlanResponse struct {
	InsurancePlanID string  `json:"insurance_plan_id"`
	PayerId         string  `json:"payer_id"`
	HisPayerId      string  `json:"his_payer_id"`
	PayerName       string  `json:"payer_name"`
	IsPrimary       bool    `json:"is_primary"`
	PatientShare    string  `json:"patient_share"`
	MaxLimit        int     `json:"max_limit"`
	RemainingLimit  float64 `json:"remaining_limit"`
}
