package controllers

import (
	"context"
	"net/http"
	"supergit/inpatient/dtos"
	"supergit/inpatient/middleware"
	"supergit/inpatient/models"
	"supergit/inpatient/utils"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func CreateSpeciality(c *gin.Context, mongoClient *mongo.Client) {
	var req dtos.CreateSpecialityReq

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

	departmentID, err := primitive.ObjectIDFromHex(req.DepartmentID)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Status:  http.StatusBadRequest,
			Message: "Invalid department ID",
		})
		return
	}

	var department models.Department
	deptFilter := bson.M{
		"_id":         departmentID,
		"business_id": user.BusinessId,
		"branch_id":   user.BranchId,
		"deleted_at":  bson.M{"$exists": false},
	}
	err = collections.Departments.FindOne(ctx, deptFilter).Decode(&department)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, utils.ErrorResponse{
				Status:  http.StatusNotFound,
				Message: "Department not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: "Failed to verify department: " + err.Error(),
		})
		return
	}

	existingFilter := bson.M{
		"business_id":   user.BusinessId,
		"branch_id":     user.BranchId,
		"department_id": departmentID,
		"deleted_at":    bson.M{"$exists": false},
		"$or": []bson.M{
			{"name_en": bson.M{"$regex": "^" + req.NameEn + "$", "$options": "i"}},
			{"name_ar": req.NameAr},
		},
	}

	var existingSpec models.Speciality
	err = collections.Specialities.FindOne(ctx, existingFilter).Decode(&existingSpec)
	if err == nil {
		c.JSON(http.StatusConflict, utils.ErrorResponse{
			Status:  http.StatusConflict,
			Message: "Speciality name already exists in this department",
		})
		return
	} else if err != mongo.ErrNoDocuments {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: "Failed to check speciality uniqueness: " + err.Error(),
		})
		return
	}

	speciality := models.Speciality{
		ID:           primitive.NewObjectID(),
		NameEn:       req.NameEn,
		NameAr:       req.NameAr,
		Description:  req.Description,
		DepartmentID: departmentID,
		BusinessID:   user.BusinessId,
		BranchID:     user.BranchId,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	_, err = collections.Specialities.InsertOne(ctx, speciality)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: "Failed to create speciality: " + err.Error(),
		})
		return
	}

	specialityRes := dtos.SpecialityRes{
		ID:           speciality.ID.Hex(),
		NameEn:       speciality.NameEn,
		NameAr:       speciality.NameAr,
		Description:  speciality.Description,
		DepartmentID: speciality.DepartmentID.Hex(),
		Department: &dtos.DepartmentBasicRes{
			ID:                  department.ID.Hex(),
			NameEn:              department.NameEn,
			NameAr:              department.NameAr,
			DepartmentShortName: department.DepartmentShortName,
		},
		BusinessID: speciality.BusinessID,
		BranchID:   speciality.BranchID,
		CreatedAt:  speciality.CreatedAt.Format(time.RFC3339),
		UpdatedAt:  speciality.UpdatedAt.Format(time.RFC3339),
	}

	c.JSON(http.StatusCreated, utils.SuccessResponse{
		Status:  http.StatusCreated,
		Message: "Speciality created successfully",
		Data: map[string]interface{}{
			"speciality": specialityRes,
		},
	})
}

func GetAllSpecialities(c *gin.Context, mongoClient *mongo.Client) {
	page := utils.StringToInt(c.DefaultQuery("page", "1"))
	perPage := utils.StringToInt(c.DefaultQuery("per_page", "10"))
	search := c.Query("search")
	departmentID := c.Query("department_id")

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
		"business_id": user.BusinessId,
		"branch_id":   user.BranchId,
		"deleted_at":  bson.M{"$exists": false},
	}

	if search != "" {
		filter["$or"] = []bson.M{
			{"name_en": bson.M{"$regex": search, "$options": "i"}},
			{"name_ar": bson.M{"$regex": search, "$options": "i"}},
		}
	}

	if departmentID != "" {
		deptObjID, err := primitive.ObjectIDFromHex(departmentID)
		if err == nil {
			filter["department_id"] = deptObjID
		}
	}

	var specialities []models.Speciality
	paginationResult, err := utils.PaginateMongo(ctx, collections.Specialities, filter, page, perPage, &specialities, "created_at", -1)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: "Failed to retrieve specialities: " + err.Error(),
		})
		return
	}

	deptIDSet := make(map[primitive.ObjectID]bool)
	for _, spec := range specialities {
		deptIDSet[spec.DepartmentID] = true
	}

	deptIDs := make([]primitive.ObjectID, 0, len(deptIDSet))
	for id := range deptIDSet {
		deptIDs = append(deptIDs, id)
	}

	var departments []models.Department
	if len(deptIDs) > 0 {
		deptFilter := bson.M{
			"_id":        bson.M{"$in": deptIDs},
			"deleted_at": bson.M{"$exists": false},
		}
		cursor, err := collections.Departments.Find(ctx, deptFilter)
		if err == nil {
			cursor.All(ctx, &departments)
			cursor.Close(ctx)
		}
	}

	deptMap := make(map[string]dtos.DepartmentBasicRes)
	for _, dept := range departments {
		deptMap[dept.ID.Hex()] = dtos.DepartmentBasicRes{
			ID:                  dept.ID.Hex(),
			NameEn:              dept.NameEn,
			NameAr:              dept.NameAr,
			DepartmentShortName: dept.DepartmentShortName,
		}
	}

	specialityList := make([]dtos.SpecialityRes, 0, len(specialities))
	for _, spec := range specialities {
		deptID := spec.DepartmentID.Hex()
		specRes := dtos.SpecialityRes{
			ID:           spec.ID.Hex(),
			NameEn:       spec.NameEn,
			NameAr:       spec.NameAr,
			Description:  spec.Description,
			DepartmentID: deptID,
			BusinessID:   spec.BusinessID,
			BranchID:     spec.BranchID,
			CreatedAt:    spec.CreatedAt.Format(time.RFC3339),
			UpdatedAt:    spec.UpdatedAt.Format(time.RFC3339),
		}

		if dept, exists := deptMap[deptID]; exists {
			specRes.Department = &dept
		}

		specialityList = append(specialityList, specRes)
	}

	c.JSON(http.StatusOK, utils.SuccessResponse{
		Status:  http.StatusOK,
		Message: "Specialities retrieved successfully",
		Data: map[string]interface{}{
			"specialities": specialityList,
			"pagination": map[string]interface{}{
				"total_records": paginationResult.TotalRecords,
				"total_pages":   paginationResult.TotalPages,
				"page":          paginationResult.Page,
				"per_page":      paginationResult.PerPage,
			},
		},
	})
}

func GetAllSpecialitiesList(c *gin.Context, mongoClient *mongo.Client) {
	search := c.Query("search")
	departmentID := c.Query("department_id")

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
		"business_id": user.BusinessId,
		"branch_id":   user.BranchId,
		"deleted_at":  bson.M{"$exists": false},
	}

	if search != "" {
		filter["$or"] = []bson.M{
			{"name_en": bson.M{"$regex": search, "$options": "i"}},
			{"name_ar": bson.M{"$regex": search, "$options": "i"}},
		}
	}

	if departmentID != "" {
		deptObjID, err := primitive.ObjectIDFromHex(departmentID)
		if err == nil {
			filter["department_id"] = deptObjID
		}
	}

	cursor, err := collections.Specialities.Find(ctx, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: "Failed to retrieve specialities",
		})
		return
	}
	defer cursor.Close(ctx)

	var specialities []models.Speciality
	if err = cursor.All(ctx, &specialities); err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: "Failed to decode specialities",
		})
		return
	}

	specialityList := make([]dtos.SpecialityRes, 0, len(specialities))
	for _, spec := range specialities {
		specialityList = append(specialityList, dtos.SpecialityRes{
			ID:           spec.ID.Hex(),
			NameEn:       spec.NameEn,
			NameAr:       spec.NameAr,
			Description:  spec.Description,
			DepartmentID: spec.DepartmentID.Hex(),
			BusinessID:   spec.BusinessID,
			BranchID:     spec.BranchID,
			CreatedAt:    spec.CreatedAt.Format(time.RFC3339),
			UpdatedAt:    spec.UpdatedAt.Format(time.RFC3339),
		})
	}

	c.JSON(http.StatusOK, utils.SuccessResponse{
		Status:  http.StatusOK,
		Message: "Specialities list retrieved successfully",
		Data: map[string]interface{}{
			"specialities": specialityList,
			"total":        len(specialityList),
		},
	})
}
func GetSpecialitiesByDepartment(c *gin.Context, mongoClient *mongo.Client) {
	departmentID := c.Param("department_id")

	if departmentID == "" {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Status:  http.StatusBadRequest,
			Message: "Department ID is required",
		})
		return
	}

	departmentObjID, err := primitive.ObjectIDFromHex(departmentID)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Status:  http.StatusBadRequest,
			Message: "Invalid department ID format",
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

	filter := bson.M{
		"department_id": departmentObjID,
		"business_id":   user.BusinessId,
		"branch_id":     user.BranchId,
		"deleted_at":    bson.M{"$exists": false},
	}

	cursor, err := collections.Specialities.Find(ctx, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: "Failed to retrieve specialities: " + err.Error(),
		})
		return
	}
	defer cursor.Close(ctx)

	var specialities []models.Speciality
	if err = cursor.All(ctx, &specialities); err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: "Failed to decode specialities: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, utils.SuccessResponse{
		Status:  http.StatusOK,
		Message: "Specialities retrieved successfully",
		Data:    specialities,
	})
}

func GetSpecialityByID(c *gin.Context, mongoClient *mongo.Client) {
	specialityID := c.Param("id")

	objectID, err := primitive.ObjectIDFromHex(specialityID)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Status:  http.StatusBadRequest,
			Message: "Invalid speciality ID",
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

	var speciality models.Speciality
	filter := bson.M{
		"_id":         objectID,
		"business_id": user.BusinessId,
		"branch_id":   user.BranchId,
		"deleted_at":  bson.M{"$exists": false},
	}

	err = collections.Specialities.FindOne(ctx, filter).Decode(&speciality)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, utils.ErrorResponse{
				Status:  http.StatusNotFound,
				Message: "Speciality not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: "Failed to retrieve speciality",
		})
		return
	}

	var department models.Department
	deptFilter := bson.M{
		"_id":        speciality.DepartmentID,
		"deleted_at": bson.M{"$exists": false},
	}
	collections.Departments.FindOne(ctx, deptFilter).Decode(&department)

	specialityRes := dtos.SpecialityRes{
		ID:           speciality.ID.Hex(),
		NameEn:       speciality.NameEn,
		NameAr:       speciality.NameAr,
		Description:  speciality.Description,
		DepartmentID: speciality.DepartmentID.Hex(),
		BusinessID:   speciality.BusinessID,
		BranchID:     speciality.BranchID,
		CreatedAt:    speciality.CreatedAt.Format(time.RFC3339),
		UpdatedAt:    speciality.UpdatedAt.Format(time.RFC3339),
	}

	if department.ID != primitive.NilObjectID {
		specialityRes.Department = &dtos.DepartmentBasicRes{
			ID:                  department.ID.Hex(),
			NameEn:              department.NameEn,
			NameAr:              department.NameAr,
			DepartmentShortName: department.DepartmentShortName,
		}
	}

	c.JSON(http.StatusOK, utils.SuccessResponse{
		Status:  http.StatusOK,
		Message: "Speciality retrieved successfully",
		Data: map[string]interface{}{
			"speciality": specialityRes,
		},
	})
}

func UpdateSpeciality(c *gin.Context, mongoClient *mongo.Client) {
	specialityID := c.Param("id")

	objectID, err := primitive.ObjectIDFromHex(specialityID)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Status:  http.StatusBadRequest,
			Message: "Invalid speciality ID",
		})
		return
	}

	var req dtos.UpdateSpecialityReq
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

	var currentSpeciality models.Speciality
	currentFilter := bson.M{
		"_id":         objectID,
		"business_id": user.BusinessId,
		"branch_id":   user.BranchId,
		"deleted_at":  bson.M{"$exists": false},
	}
	err = collections.Specialities.FindOne(ctx, currentFilter).Decode(&currentSpeciality)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, utils.ErrorResponse{
				Status:  http.StatusNotFound,
				Message: "Speciality not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: "Failed to retrieve speciality: " + err.Error(),
		})
		return
	}

	checkDepartmentID := currentSpeciality.DepartmentID
	if req.DepartmentID != "" {
		deptObjID, err := primitive.ObjectIDFromHex(req.DepartmentID)
		if err != nil {
			c.JSON(http.StatusBadRequest, utils.ErrorResponse{
				Status:  http.StatusBadRequest,
				Message: "Invalid department ID",
			})
			return
		}
		checkDepartmentID = deptObjID
		var department models.Department
		deptFilter := bson.M{
			"_id":         deptObjID,
			"business_id": user.BusinessId,
			"branch_id":   user.BranchId,
			"deleted_at":  bson.M{"$exists": false},
		}
		err = collections.Departments.FindOne(ctx, deptFilter).Decode(&department)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				c.JSON(http.StatusNotFound, utils.ErrorResponse{
					Status:  http.StatusNotFound,
					Message: "Department not found",
				})
				return
			}
			c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
				Status:  http.StatusInternalServerError,
				Message: "Failed to verify department: " + err.Error(),
			})
			return
		}
	}

	if req.NameEn != "" || req.NameAr != "" {
		existingFilter := bson.M{
			"business_id":   user.BusinessId,
			"branch_id":     user.BranchId,
			"department_id": checkDepartmentID,
			"deleted_at":    bson.M{"$exists": false},
			"_id":           bson.M{"$ne": objectID},
		}

		orConditions := []bson.M{}
		if req.NameEn != "" {
			orConditions = append(orConditions, bson.M{"name_en": bson.M{"$regex": "^" + req.NameEn + "$", "$options": "i"}})
		}
		if req.NameAr != "" {
			orConditions = append(orConditions, bson.M{"name_ar": req.NameAr})
		}

		if len(orConditions) > 0 {
			existingFilter["$or"] = orConditions
		}

		var existingSpec models.Speciality
		err := collections.Specialities.FindOne(ctx, existingFilter).Decode(&existingSpec)
		if err == nil {
			c.JSON(http.StatusConflict, utils.ErrorResponse{
				Status:  http.StatusConflict,
				Message: "Speciality name already exists in this department",
			})
			return
		} else if err != mongo.ErrNoDocuments {
			c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
				Status:  http.StatusInternalServerError,
				Message: "Failed to check speciality uniqueness: " + err.Error(),
			})
			return
		}
	}

	filter := bson.M{
		"_id":         objectID,
		"business_id": user.BusinessId,
		"branch_id":   user.BranchId,
		"deleted_at":  bson.M{"$exists": false},
	}

	update := bson.M{
		"$set": bson.M{
			"updated_at": time.Now(),
		},
	}

	if req.NameEn != "" {
		update["$set"].(bson.M)["name_en"] = req.NameEn
	}
	if req.NameAr != "" {
		update["$set"].(bson.M)["name_ar"] = req.NameAr
	}
	if req.Description != "" {
		update["$set"].(bson.M)["description"] = req.Description
	}
	if req.DepartmentID != "" {
		deptObjID, _ := primitive.ObjectIDFromHex(req.DepartmentID)
		update["$set"].(bson.M)["department_id"] = deptObjID
	}

	result, err := collections.Specialities.UpdateOne(ctx, filter, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: "Failed to update speciality: " + err.Error(),
		})
		return
	}

	if result.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, utils.ErrorResponse{
			Status:  http.StatusNotFound,
			Message: "Speciality not found",
		})
		return
	}

	var updatedSpeciality models.Speciality
	err = collections.Specialities.FindOne(ctx, filter).Decode(&updatedSpeciality)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: "Failed to retrieve updated speciality",
		})
		return
	}

	var department models.Department
	deptFilter := bson.M{
		"_id":        updatedSpeciality.DepartmentID,
		"deleted_at": bson.M{"$exists": false},
	}
	collections.Departments.FindOne(ctx, deptFilter).Decode(&department)

	specialityRes := dtos.SpecialityRes{
		ID:           updatedSpeciality.ID.Hex(),
		NameEn:       updatedSpeciality.NameEn,
		NameAr:       updatedSpeciality.NameAr,
		Description:  updatedSpeciality.Description,
		DepartmentID: updatedSpeciality.DepartmentID.Hex(),
		BusinessID:   updatedSpeciality.BusinessID,
		BranchID:     updatedSpeciality.BranchID,
		CreatedAt:    updatedSpeciality.CreatedAt.Format(time.RFC3339),
		UpdatedAt:    updatedSpeciality.UpdatedAt.Format(time.RFC3339),
	}

	if department.ID != primitive.NilObjectID {
		specialityRes.Department = &dtos.DepartmentBasicRes{
			ID:                  department.ID.Hex(),
			NameEn:              department.NameEn,
			NameAr:              department.NameAr,
			DepartmentShortName: department.DepartmentShortName,
		}
	}

	c.JSON(http.StatusOK, utils.SuccessResponse{
		Status:  http.StatusOK,
		Message: "Speciality updated successfully",
		Data: map[string]interface{}{
			"speciality": specialityRes,
		},
	})
}

func DeleteSpeciality(c *gin.Context, mongoClient *mongo.Client) {
	specialityID := c.Param("id")

	objectID, err := primitive.ObjectIDFromHex(specialityID)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Status:  http.StatusBadRequest,
			Message: "Invalid speciality ID",
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

	filter := bson.M{
		"_id":         objectID,
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

	result, err := collections.Specialities.UpdateOne(ctx, filter, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: "Failed to delete speciality: " + err.Error(),
		})
		return
	}

	if result.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, utils.ErrorResponse{
			Status:  http.StatusNotFound,
			Message: "Speciality not found",
		})
		return
	}

	c.JSON(http.StatusOK, utils.SuccessResponse{
		Status:  http.StatusOK,
		Message: "Speciality deleted successfully",
		Data:    nil,
	})
}
