package models

type Payer struct {
	ID                 string             `bson:"_id,omitempty" json:"id,omitempty"`
	PayerID            string             `bson:"payer_id" json:"payer_id"`
	RCMRef             string             `bson:"rcm_ref" json:"rcm_ref"`
	ErpRef             any                `bson:"erp_ref" json:"erp_ref"`
	PayerName          string             `bson:"name" json:"name"`
	PharmacyDetails    PharmacyDetails    `bson:"pharmacy_details" json:"pharmacy_details"`
	Generals           Generals           `bson:"generals" json:"generals,omitempty"`
	PatientShare       string             `bson:"patient_share" json:"patient_share"`
	MaxLimit           int                `bson:"max_limit" json:"max_limit"`
	NameAr             string             `bson:"name_ar" json:"name_ar"`
	NameEn             string             `bson:"name_en" json:"name_en"`
	TpaNameEn          string             `bson:"tpa_name_en" json:"tpa_name_en"`
	TpaNameAr          string             `bson:"tpa_name_ar" json:"tpa_name_ar"`
	BusinessID         uint               `bson:"business_id" json:"business_id"`
	IsMasterPrice      bool               `bson:"is_master_price" json:"is_master_price"`
	BranchID           uint               `bson:"branch_id" json:"branch_id"`
	LicenseID          string             `bson:"license_id" json:"license_id"`
	TpaLicenseID       string             `bson:"tpa_license_id" json:"tpa_license_id"`
	Type               string             `bson:"type" json:"type"`
	Category           string             `bson:"category" json:"category"`
	ProviderVatNo      uint               `bson:"provider_vat_no" json:"provider_vat_no"`
	PayarVatNo         uint               `bson:"payar_vat_no" json:"payar_vat_no"`
	PayarCRNo          uint               `bson:"payar_cr_no" json:"payar_cr_no"`
	PayarLogo          string             `bson:"payar_logo" json:"payar_logo"`
	PayarAddress       PayarAddress       `bson:"payar_address" json:"payar_address"`
	AccountManager     AccountManager     `bson:"account_manager" json:"account_manager"`
	FaxNo              string             `bson:"fax_no" json:"fax_no"`
	TelephoneNumber    string             `bson:"telephone_number" json:"telephone_number"`
	DiscountParcentage float64            `bson:"discount_parcentage" json:"discount_parcentage"`
	ContractStarteDate string             `bson:"contract_start_date" json:"contract_start_date"`
	ContractEndeDate   string             `bson:"contract_end_date" json:"contract_end_date"`
	IsVatApplicable    bool               `bson:"is_vat_applicable" json:"is_vat_applicable"`
	ServiceDiscount    []ServiceDiscount  `bson:"service_discount" json:"service_discount"`
}

type PayerInfo struct {
	ID              string `bson:"_id,omitempty" json:"id,omitempty"`
	PayerID         string `bson:"payer_id" json:"payer_id"`
	HisPayerID      string `bson:"his_payer_id" json:"his_payer_id"`
	RCMRef          string `bson:"rcm_ref" json:"rcm_ref"`
	ErpRef          any    `bson:"erp_ref" json:"erp_ref"`
	PayerName       string `bson:"name" json:"name"`
	NameAr          string `bson:"name_ar" json:"name_ar"`
	NameEn          string `bson:"name_en" json:"name_en"`
	TpaNameEn       string `bson:"tpa_name_en" json:"tpa_name_en"`
	TpaNameAr       string `bson:"tpa_name_ar" json:"tpa_name_ar"`
	LicenseID       string `bson:"license_id" json:"license_id"`
	TpaLicenseID    string `bson:"tpa_license_id" json:"tpa_license_id"`
	Type            string `bson:"type" json:"type"`
	Category        string `bson:"category" json:"category"`
	ProviderVatNo   uint   `bson:"provider_vat_no" json:"provider_vat_no"`
	PayarVatNo      uint   `bson:"payar_vat_no" json:"payar_vat_no"`
	PayarCRNo       uint   `bson:"payar_cr_no" json:"payar_cr_no"`
	IsVatApplicable bool   `bson:"is_vat_applicable" json:"is_vat_applicable"`
}

type SettingDetails struct {
	PharmacyDiscount      float64 `bson:"pharmacy_discount" json:"pharmacy_discount"`
	PharmacyApprovalLimit float64 `bson:"pharmacy_approval_limit" json:"pharmacy_approval_limit"`
}

type NewSettingDetails struct {
	PharmacyDeductable    float64 `bson:"pharmacy_deductable" json:"pharmacy_deductable"`
	PharmacyApprovalLimit float64 `bson:"pharmacy_approval_limit" json:"pharmacy_approval_limit"`
}

type PharmacyDetails struct {
	Generics         SettingDetails `bson:"generics" json:"generics,omitempty"`
	BrandReplaceable SettingDetails `bson:"brand_replaceable" json:"brand_replaceable,omitempty"`
	MedicalDevice    SettingDetails `bson:"medical_device" json:"medical_device,omitempty"`
	Milk             SettingDetails `bson:"milk" json:"milk,omitempty"`
}

type ServiceDiscount struct {
	ServiceID       string  `bson:"service_id" json:"service_id"`
	Patientshare    float64 `bson:"patient_share" json:"patient_share"`
	PatientDiscount float64 `bson:"patient_discount" json:"patient_discount"`
}

type PharmacyPoliciesDeductable struct {
	Generics         NewSettingDetails `bson:"generics" json:"generics,omitempty"`
	BrandReplaceable NewSettingDetails `bson:"brand_replaceable" json:"brand_replaceable,omitempty"`
	MedicalDevice    NewSettingDetails `bson:"medical_device" json:"medical_device,omitempty"`
	Milk             NewSettingDetails `bson:"milk" json:"milk,omitempty"`
}

type PayarAddress struct {
	Street                 string `bson:"street" json:"street"`
	BuildingNumber         string `bson:"building_no" json:"building_no"`
	City                   string `bson:"city" json:"city"`
	PostalCode             string `bson:"postal_code" json:"postal_code"`
	StateOrProvince        string `bson:"state_or_province" json:"state_or_province"`
	DistrictOrNeighborhood string `bson:"district_or_neighborhood" json:"district_or_neighborhood"`
}

type AccountManager struct {
	NameEn   string `bson:"name_en" json:"name_en"`
	NameAr   string `bson:"name_ar" json:"name_ar"`
	MobileNo string `bson:"mobile_no" json:"mobile_no"`
	Email    string `bson:"email" json:"email"`
}

type Generals struct {
	Deductable    *float64 `bson:"deductable,omitempty" json:"deductable,omitempty"`
	Discount      *float64 `bson:"discount,omitempty" json:"discount,omitempty"`
	ApprovalLimit *float64 `bson:"approval_limit,omitempty" json:"approval_limit,omitempty"`
}
