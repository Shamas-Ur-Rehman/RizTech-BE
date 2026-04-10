package controllers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"supergit/inpatient/config"
	"supergit/inpatient/dtos"
	"supergit/inpatient/middleware"
	"supergit/inpatient/models"
	"supergit/inpatient/utils"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func CreatePatient(c *gin.Context, mongoClient *mongo.Client) {
	var req dtos.CreatePatientDto

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Status:  http.StatusBadRequest,
			Message: "Invalid request: " + err.Error(),
		})
		return
	}
	validationErrors := utils.ValidatePatientData(req.FullName, req.DocumentID, req.FileNumber, req.Email, req.Contact, req.Gender, req.MartialStatus)
	if len(validationErrors) > 0 {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Status:  http.StatusBadRequest,
			Message: utils.CombineValidationErrors(validationErrors),
		})
		return
	}

	collections := middleware.GetCollections(c)
	if collections == nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: "Collections not found in context",
		})
		return
	}
	user := middleware.GetUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, utils.ErrorResponse{
			Status:  http.StatusUnauthorized,
			Message: "User not found in context",
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	existingFilter := bson.M{
		"business_id": user.BusinessId,
		"branch_id":   user.BranchId,
		"deleted_at":  bson.M{"$exists": false},
		"$or": []bson.M{
			{"document_id": req.DocumentID},
		},
	}
	if req.FileNumber != "" {
		existingFilter["$or"] = append(existingFilter["$or"].([]bson.M), bson.M{"file_no": req.FileNumber})
	}
	var existingPatient models.Patient
	err := collections.Patients.FindOne(ctx, existingFilter).Decode(&existingPatient)
	if err == nil {
		if existingPatient.DocumentID == req.DocumentID {
			c.JSON(http.StatusConflict, utils.ErrorResponse{
				Status:  http.StatusConflict,
				Message: "Patient with this document ID already exists",
			})
			return
		}
		if existingPatient.FileNumber == req.FileNumber && req.FileNumber != "" {
			c.JSON(http.StatusConflict, utils.ErrorResponse{
				Status:  http.StatusConflict,
				Message: "Patient with this file number already exists",
			})
			return
		}
	} else if err != mongo.ErrNoDocuments {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: "Failed to check for duplicate patient: " + err.Error(),
		})
		return
	}
	patientID, err := generatePatientID(ctx, collections, user.BusinessId, user.BranchId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: "Failed to generate patient ID: " + err.Error(),
		})
		return
	}

	insurancePlans := make([]models.InsurancePlan, len(req.InsurancePlan))
	for i, plan := range req.InsurancePlan {
		insurancePlans[i] = models.InsurancePlan{
			InsurancePlanID:        plan.InsurancePlanID,
			MemberCardId:           plan.MemberCardId,
			PolicyNumber:           plan.PolicyNumber,
			ExpiryDate:             plan.ExpiryDate,
			IsPrimary:              plan.IsPrimary,
			PayerId:                plan.PayerId,
			HisPayerId:             plan.HisPayerId,
			PayerName:              plan.PayerName,
			RelationWithSubscriber: plan.RelationWithSubscriber,
			CoverageType:           plan.CoverageType,
			PatientShare:           plan.PatientShare,
			MaxLimit:               plan.MaxLimit,
			DiscountPercentage:     plan.DiscountPercentage,
			RemainingLimit:         plan.RemainingLimit,
			Network:                plan.Network,
			IssueDate:              plan.IssueDate,
			SponsorNo:              plan.SponsorNo,
			PolicyHolder:           plan.PolicyHolder,
			InsuranceType:          plan.InsuranceType,
			InsuranceStatus:        plan.InsuranceStatus,
			InsuranceDuration:      plan.InsuranceDuration,
			ClassID:                plan.ClassID,
			ClassName:              plan.ClassName,
			ClassType:              plan.ClassType,
			PolicyClass:            plan.PolicyClass,
			PolicyClassID:          plan.PolicyClassID,
			PolicyClassName:        plan.PolicyClassName,
			PolicyClassType:        plan.PolicyClassType,
		}
	}
	subscriberID := req.SubscriberID

	subscriberRelationship := req.SubscriberRelationship

	subscriberInsurancePlan := req.SubscriberInsurancePlan

	contactNumber := req.ContactNumber
	if contactNumber == "" {
		contactNumber = req.Contact
	}

	patient := models.Patient{
		PatientID:               patientID,
		OutPatientID:            req.OutPatientID,
		FullName:                req.FullName,
		Contact:                 req.Contact,
		ContactNumber:           contactNumber,
		DocumentID:              req.DocumentID,
		DocumentType:            req.DocumentType,
		Gender:                  req.Gender,
		BirthDate:               req.BirthDate,
		DateType:                req.DateType,
		Email:                   req.Email,
		FileNumber:              req.FileNumber,
		RCMRef:                  req.RCMRef,
		PatientType:             req.PatientType,
		Nationality:             req.Nationality,
		Address:                 req.Address,
		City:                    req.City,
		BloodGroup:              req.BloodGroup,
		MartialStatus:           req.MartialStatus,
		Occupation:              req.Occupation,
		Religion:                req.Religion,
		ResidencyType:           req.ResidencyType,
		PassportNumber:          req.PassportNo,
		VisaTitle:               req.VisaTitle,
		VisaNumber:              req.VisaNo,
		VisaType:                req.VisaType,
		BorderNumber:            req.BorderNo,
		InsuranceDuration:       req.InsuranceDuration,
		IsNewBorn:               req.IsNewBorn,
		Role:                    req.Role,
		SubscriberID:            subscriberID,
		SubscriberRelationship:  subscriberRelationship,
		SubscriberInsurancePlan: subscriberInsurancePlan,
		InsurancePlan:           insurancePlans,
		Guarantors:              req.Guarantors,
		EmergencyContacts:       req.EmergencyContacts,
		PragnancyHistory:        req.PragnancyHistory,
		BusinessID:              user.BusinessId,
		BranchID:                user.BranchId,
		CreatedAt:               time.Now(),
		UpdatedAt:               time.Now(),
	}

	result, err := collections.Patients.InsertOne(ctx, patient)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: "Failed to create patient: " + err.Error(),
		})
		return
	}
	c.JSON(http.StatusCreated, utils.SuccessResponse{
		Status:  http.StatusCreated,
		Message: "Patient created successfully",
		Data: map[string]interface{}{
			"patient_id": patient.PatientID,
			"_id":        result.InsertedID,
		},
	})
}

func GetAllPatients(c *gin.Context, mongoClient *mongo.Client) {
	collections := middleware.GetCollections(c)
	if collections == nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: "Collections not found in context",
		})
		return
	}
	user := middleware.GetUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, utils.ErrorResponse{
			Status:  http.StatusUnauthorized,
			Message: "User not found in context",
		})
		return
	}
	page := utils.StringToInt(c.DefaultQuery("page", "1"))
	perPage := utils.StringToInt(c.DefaultQuery("per_page", "10"))
	sortField := c.DefaultQuery("sort_field", "createdAt")
	sortOrder := utils.StringToInt(c.DefaultQuery("sort_order", "-1"))

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"business_id": user.BusinessId,
		"branch_id":   user.BranchId,
		"deleted_at":  bson.M{"$exists": false},
	}

	var patients []models.Patient
	paginationResult, err := utils.PaginateMongo(ctx, collections.Patients, filter, page, perPage, &patients, sortField, sortOrder)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: "Failed to retrieve patients: " + err.Error(),
		})
		return
	}
	patientList := make([]dtos.PatientResponseDto, 0)
	for _, patient := range patients {
		patientList = append(patientList, mapToResponseDto(patient))
	}
	c.JSON(http.StatusOK, utils.SuccessResponse{
		Status:  http.StatusOK,
		Message: "Patients retrieved successfully from MongoDB",
		Data: map[string]interface{}{
			"patients": patientList,
			"pagination": map[string]interface{}{
				"total_records": paginationResult.TotalRecords,
				"total_pages":   paginationResult.TotalPages,
				"page":          paginationResult.Page,
				"per_page":      paginationResult.PerPage,
			},
		},
	})
}

func GetPatientByID(c *gin.Context, mongoClient *mongo.Client) {
	patientID := c.Param("id")

	collections := middleware.GetCollections(c)
	if collections == nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: "Collections not found in context",
		})
		return
	}
	user := middleware.GetUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, utils.ErrorResponse{
			Status:  http.StatusUnauthorized,
			Message: "User not found in context",
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"patient_id":  patientID,
		"business_id": user.BusinessId,
		"branch_id":   user.BranchId,
		"deleted_at":  bson.M{"$exists": false},
	}

	var patient models.Patient
	if err := collections.Patients.FindOne(ctx, filter).Decode(&patient); err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, utils.ErrorResponse{
				Status:  http.StatusNotFound,
				Message: "Patient not found or access denied",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: "Failed to retrieve patient: " + err.Error(),
		})
		return
	}

	patientRes := mapToResponseDto(patient)

	c.JSON(http.StatusOK, utils.SuccessResponse{
		Status:  http.StatusOK,
		Message: "Patient retrieved successfully from MongoDB",
		Data: map[string]interface{}{
			"patient": patientRes,
		},
	})
}

func UpdatePatient(c *gin.Context, mongoClient *mongo.Client) {
	patientID := c.Param("id")

	var req dtos.UpdatePatientDto

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Status:  http.StatusBadRequest,
			Message: "Invalid request: " + err.Error(),
		})
		return
	}

	collections := middleware.GetCollections(c)
	if collections == nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: "Collections not found in context",
		})
		return
	}

	user := middleware.GetUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, utils.ErrorResponse{
			Status:  http.StatusUnauthorized,
			Message: "User not found in context",
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	update := bson.M{
		"$set": bson.M{
			"updated_at": time.Now(),
		},
	}

	if req.FullName != "" {
		update["$set"].(bson.M)["full_name"] = req.FullName
	}
	if req.Contact != "" {
		update["$set"].(bson.M)["contact"] = req.Contact
	}
	if req.ContactNumber != "" {
		update["$set"].(bson.M)["contact_number"] = req.ContactNumber
	}
	if req.DocumentID != "" {
		update["$set"].(bson.M)["document_id"] = req.DocumentID
	}
	if req.DocumentType != "" {
		update["$set"].(bson.M)["document_type"] = req.DocumentType
	}
	if req.Gender != "" {
		update["$set"].(bson.M)["gender"] = req.Gender
	}
	if req.BirthDate != "" {
		update["$set"].(bson.M)["dob"] = req.BirthDate
	}
	if req.DateType != "" {
		update["$set"].(bson.M)["date_type"] = req.DateType
	}
	if req.Email != "" {
		update["$set"].(bson.M)["email"] = req.Email
	}
	if req.FileNumber != "" {
		update["$set"].(bson.M)["file_no"] = req.FileNumber
	}
	if req.PatientID != "" {
		update["$set"].(bson.M)["patient_id"] = req.PatientID
	}
	if req.RCMRef != "" {
		update["$set"].(bson.M)["rcm_ref"] = req.RCMRef
	}
	if req.PatientType != "" {
		update["$set"].(bson.M)["beneficiary_type"] = req.PatientType
	}
	if req.Nationality != "" {
		update["$set"].(bson.M)["nationality"] = req.Nationality
	}
	if req.Address != "" {
		update["$set"].(bson.M)["address"] = req.Address
	}
	if req.City != "" {
		update["$set"].(bson.M)["city"] = req.City
	}
	if req.BloodGroup != "" {
		update["$set"].(bson.M)["blood_group"] = req.BloodGroup
	}
	if req.MartialStatus != "" {
		update["$set"].(bson.M)["martial_status"] = req.MartialStatus
	}
	if req.Occupation != "" {
		update["$set"].(bson.M)["occupation"] = req.Occupation
	}
	if req.Religion != "" {
		update["$set"].(bson.M)["religion"] = req.Religion
	}
	if req.ResidencyType != "" {
		update["$set"].(bson.M)["residency_type"] = req.ResidencyType
	}
	if req.PassportNo != "" {
		update["$set"].(bson.M)["passport_no"] = req.PassportNo
	}
	if req.VisaTitle != "" {
		update["$set"].(bson.M)["visa_title"] = req.VisaTitle
	}
	if req.VisaNo != "" {
		update["$set"].(bson.M)["visa_no"] = req.VisaNo
	}
	if req.VisaType != "" {
		update["$set"].(bson.M)["visa_type"] = req.VisaType
	}
	if req.BorderNo != "" {
		update["$set"].(bson.M)["border_no"] = req.BorderNo
	}
	if req.InsuranceDuration != "" {
		update["$set"].(bson.M)["insurance_duration"] = req.InsuranceDuration
	}
	if req.SubscriberID != "" {
		update["$set"].(bson.M)["subscriber_id"] = req.SubscriberID
	}
	if req.SubscriberRelationship != "" {
		update["$set"].(bson.M)["subscriber_relationship"] = req.SubscriberRelationship
	}
	if len(req.SubscriberInsurancePlan) > 0 {
		update["$set"].(bson.M)["subscriber_insurance_plan"] = req.SubscriberInsurancePlan
	}
	if req.SubscriberIdAlt != "" {
		update["$set"].(bson.M)["subscriberId"] = req.SubscriberIdAlt
	}
	if req.SubscriberRelationshipAlt != "" {
		update["$set"].(bson.M)["subscriberRelationship"] = req.SubscriberRelationshipAlt
	}
	if len(req.SubscriberInsurancePlanAlt) > 0 {
		update["$set"].(bson.M)["subscriberInsurancePlan"] = req.SubscriberInsurancePlanAlt
	}
	if len(req.InsurancePlan) > 0 {
		insurancePlans := make([]models.InsurancePlan, len(req.InsurancePlan))
		for i, plan := range req.InsurancePlan {
			insurancePlans[i] = models.InsurancePlan{
				InsurancePlanID:        plan.InsurancePlanID,
				MemberCardId:           plan.MemberCardId,
				PolicyNumber:           plan.PolicyNumber,
				ExpiryDate:             plan.ExpiryDate,
				IsPrimary:              plan.IsPrimary,
				PayerId:                plan.PayerId,
				HisPayerId:             plan.HisPayerId,
				PayerName:              plan.PayerName,
				RelationWithSubscriber: plan.RelationWithSubscriber,
				CoverageType:           plan.CoverageType,
				PatientShare:           plan.PatientShare,
				MaxLimit:               plan.MaxLimit,
				DiscountPercentage:     plan.DiscountPercentage,
				RemainingLimit:         plan.RemainingLimit,
				Network:                plan.Network,
				IssueDate:              plan.IssueDate,
				SponsorNo:              plan.SponsorNo,
				PolicyHolder:           plan.PolicyHolder,
				InsuranceType:          plan.InsuranceType,
				InsuranceStatus:        plan.InsuranceStatus,
				InsuranceDuration:      plan.InsuranceDuration,
				ClassID:                plan.ClassID,
				ClassName:              plan.ClassName,
				ClassType:              plan.ClassType,
				PolicyClass:            plan.PolicyClass,
				PolicyClassID:          plan.PolicyClassID,
				PolicyClassName:        plan.PolicyClassName,
				PolicyClassType:        plan.PolicyClassType,
			}
		}
		update["$set"].(bson.M)["insurance_plans"] = insurancePlans
	}
	if len(req.Guarantors) > 0 {
		update["$set"].(bson.M)["guarantors"] = req.Guarantors
	}
	if len(req.EmergencyContacts) > 0 {
		update["$set"].(bson.M)["emergency_contacts"] = req.EmergencyContacts
	}
	if len(req.PragnancyHistory) > 0 {
		update["$set"].(bson.M)["pragnancy_history"] = req.PragnancyHistory
	}

	filter := bson.M{
		"patient_id":  patientID,
		"business_id": user.BusinessId,
		"branch_id":   user.BranchId,
		"deleted_at":  bson.M{"$exists": false},
	}

	result, err := collections.Patients.UpdateOne(ctx, filter, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: "Failed to update patient: " + err.Error(),
		})
		return
	}

	if result.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, utils.ErrorResponse{
			Status:  http.StatusNotFound,
			Message: "Patient not found or access denied",
		})
		return
	}

	var patient models.Patient
	if err := collections.Patients.FindOne(ctx, filter).Decode(&patient); err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: "Failed to retrieve updated patient: " + err.Error(),
		})
		return
	}

	patientRes := mapToResponseDto(patient)

	c.JSON(http.StatusOK, utils.SuccessResponse{
		Status:  http.StatusOK,
		Message: "Patient updated successfully in MongoDB",
		Data: map[string]interface{}{
			"patient": patientRes,
		},
	})
}

func DeletePatient(c *gin.Context, mongoClient *mongo.Client) {
	patientID := c.Param("id")

	collections := middleware.GetCollections(c)
	if collections == nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: "Collections not found in context",
		})
		return
	}

	user := middleware.GetUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, utils.ErrorResponse{
			Status:  http.StatusUnauthorized,
			Message: "User not found in context",
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"patient_id":  patientID,
		"business_id": user.BusinessId,
		"branch_id":   user.BranchId,
		"deleted_at":  bson.M{"$exists": false},
	}
	now := time.Now()
	update := bson.M{
		"$set": bson.M{
			"deleted_at": now,
			"updated_at": now,
		},
	}
	result, err := collections.Patients.UpdateOne(ctx, filter, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: "Failed to delete patient: " + err.Error(),
		})
		return
	}

	if result.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, utils.ErrorResponse{
			Status:  http.StatusNotFound,
			Message: "Patient not found or access denied",
		})
		return
	}

	c.JSON(http.StatusOK, utils.SuccessResponse{
		Status:  http.StatusOK,
		Message: "Patient deleted successfully from MongoDB",
		Data:    map[string]interface{}{},
	})
}
func mapToResponseDto(patient models.Patient) dtos.PatientResponseDto {
	insurancePlans := make([]dtos.InsurancePlanResponse, len(patient.InsurancePlan))
	for i, plan := range patient.InsurancePlan {
		insurancePlans[i] = dtos.InsurancePlanResponse{
			InsurancePlanID: plan.InsurancePlanID,
			PayerId:         plan.PayerId,
			HisPayerId:      plan.HisPayerId,
			PayerName:       plan.PayerName,
			IsPrimary:       plan.IsPrimary,
			PatientShare:    plan.PatientShare,
			MaxLimit:        plan.MaxLimit,
			RemainingLimit:  plan.RemainingLimit,
		}
	}
	return dtos.PatientResponseDto{
		PatientID:         patient.PatientID,
		OutPatientID:      patient.OutPatientID,
		FullName:          patient.FullName,
		Gender:            patient.Gender,
		DocumentID:        patient.DocumentID,
		PatientType:       patient.PatientType,
		FileNumber:        patient.FileNumber,
		Nationality:       patient.Nationality,
		Address:           patient.Address,
		ERPRef:            patient.ERPRef,
		BirthDate:         patient.BirthDate,
		Contact:           patient.Contact,
		Email:             patient.Email,
		Age:               patient.Age(),
		BusinessID:        patient.BusinessID,
		BranchID:          patient.BranchID,
		InsurancePlan:     insurancePlans,
		Guarantors:        patient.Guarantors,
		EmergencyContacts: patient.EmergencyContacts,
		CreatedAt:         patient.CreatedAt.Format(time.RFC3339),
		UpdatedAt:         patient.UpdatedAt.Format(time.RFC3339),
	}
}

func generatePatientID(ctx context.Context, collections *config.Collections, businessID, branchID uint) (string, error) {
	const maxRetries = 10
	const maxPatientNumber = 9999999

	for attempt := 0; attempt < maxRetries; attempt++ {
		filter := bson.M{
			"business_id": businessID,
			"branch_id":   branchID,
			"patient_id":  bson.M{"$regex": "^IP-[0-9]+$"},
		}

		pipeline := mongo.Pipeline{
			{{Key: "$match", Value: filter}},
			{{Key: "$sort", Value: bson.D{{Key: "patient_id", Value: -1}}}},
			{{Key: "$limit", Value: 1}},
			{{Key: "$project", Value: bson.D{{Key: "patient_id", Value: 1}}}},
		}

		cursor, err := collections.Patients.Aggregate(ctx, pipeline)
		if err != nil {
			return "", fmt.Errorf("failed to query patients: %w", err)
		}

		var results []models.Patient
		if err = cursor.All(ctx, &results); err != nil {
			cursor.Close(ctx)
			return "", fmt.Errorf("failed to decode patients: %w", err)
		}
		cursor.Close(ctx)

		var nextNumber int
		if len(results) == 0 {
			nextNumber = 1
		} else {
			lastPatient := results[0]
			var lastNumber int
			_, scanErr := fmt.Sscanf(lastPatient.PatientID, "IP-%d", &lastNumber)
			if scanErr == nil {
				nextNumber = lastNumber + 1
			} else {
				count, countErr := collections.Patients.CountDocuments(ctx, filter)
				if countErr != nil {
					return "", fmt.Errorf("failed to count patients: %w", countErr)
				}
				nextNumber = int(count) + 1
			}
		}
		if nextNumber > maxPatientNumber {
			return "", fmt.Errorf("patient ID limit exceeded: maximum %d patients per business/branch (enterprise limit)", maxPatientNumber)
		}
		var newPatientID string
		if nextNumber <= 999999 {
			newPatientID = fmt.Sprintf("IP-%06d", nextNumber)
		} else {
			newPatientID = fmt.Sprintf("IP-%07d", nextNumber)
		}
		checkFilter := bson.M{
			"patient_id":  newPatientID,
			"business_id": businessID,
			"branch_id":   branchID,
		}
		var existingPatient models.Patient
		err = collections.Patients.FindOne(ctx, checkFilter).Decode(&existingPatient)
		if err == mongo.ErrNoDocuments {
			return newPatientID, nil
		} else if err != nil {
			return "", fmt.Errorf("failed to check patient ID uniqueness: %w", err)
		}
		backoffTime := time.Millisecond * time.Duration(10*(attempt+1))
		time.Sleep(backoffTime)
	}
	return "", fmt.Errorf("failed to generate unique patient ID after %d attempts - high concurrency detected, please retry", maxRetries)
}
