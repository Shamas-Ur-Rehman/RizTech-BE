package controllers

import (
	"context"
	"net/http"
	"time"
	"supergit/inpatient/dtos"
	"supergit/inpatient/middleware"
	"supergit/inpatient/models"
	"supergit/inpatient/utils"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func CreateDepartment(c *gin.Context, mongoClient *mongo.Client) {
	var req dtos.CreateDepartmentReq

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

	existingFilter := bson.M{
		"business_id": user.BusinessId,
		"branch_id":   user.BranchId,
		"deleted_at":  bson.M{"$exists": false},
		"$or": []bson.M{
			{"name_en": bson.M{"$regex": "^" + req.NameEn + "$", "$options": "i"}},
			{"name_ar": req.NameAr},
		},
	}

	var existingDept models.Department
	err := collections.Departments.FindOne(ctx, existingFilter).Decode(&existingDept)
	if err == nil {
		c.JSON(http.StatusConflict, utils.ErrorResponse{
			Status:  http.StatusConflict,
			Message: "Department name already exists",
		})
		return
	} else if err != mongo.ErrNoDocuments {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: "Failed to check department uniqueness: " + err.Error(),
		})
		return
	}

	department := models.Department{
		ID:                  primitive.NewObjectID(),
		NameEn:              req.NameEn,
		NameAr:              req.NameAr,
		DescriptionEn:       req.DescriptionEn,
		DescriptionAr:       req.DescriptionAr,
		DepartmentShortName: req.DepartmentShortName,
		ERPBranchID:         req.ERPBranchID,
		IsCovered:           req.IsCovered,
		BusinessID:          user.BusinessId,
		BranchID:            user.BranchId,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}

	_, err = collections.Departments.InsertOne(ctx, department)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: "Failed to create department: " + err.Error(),
		})
		return
	}

	departmentRes := dtos.DepartmentRes{
		ID:                  department.ID.Hex(),
		NameEn:              department.NameEn,
		NameAr:              department.NameAr,
		DescriptionEn:       department.DescriptionEn,
		DescriptionAr:       department.DescriptionAr,
		DepartmentShortName: department.DepartmentShortName,
		ERPBranchID:         department.ERPBranchID,
		IsCovered:           department.IsCovered,
		BusinessID:          department.BusinessID,
		BranchID:            department.BranchID,
		CreatedAt:           department.CreatedAt.Format(time.RFC3339),
		UpdatedAt:           department.UpdatedAt.Format(time.RFC3339),
	}

	c.JSON(http.StatusCreated, utils.SuccessResponse{
		Status:  http.StatusCreated,
		Message: "Department created successfully",
		Data: map[string]interface{}{
			"department": departmentRes,
		},
	})
}

func GetAllDepartments(c *gin.Context, mongoClient *mongo.Client) {
	page := utils.StringToInt(c.DefaultQuery("page", "1"))
	perPage := utils.StringToInt(c.DefaultQuery("per_page", "10"))
	search := c.Query("search")

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
			{"department_short_name": bson.M{"$regex": search, "$options": "i"}},
		}
	}

	var departments []models.Department
	paginationResult, err := utils.PaginateMongo(ctx, collections.Departments, filter, page, perPage, &departments, "created_at", -1)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: "Failed to retrieve departments: " + err.Error(),
		})
		return
	}

	departmentList := make([]dtos.DepartmentRes, 0, len(departments))
	for _, dept := range departments {
		departmentList = append(departmentList, dtos.DepartmentRes{
			ID:                  dept.ID.Hex(),
			NameEn:              dept.NameEn,
			NameAr:              dept.NameAr,
			DescriptionEn:       dept.DescriptionEn,
			DescriptionAr:       dept.DescriptionAr,
			DepartmentShortName: dept.DepartmentShortName,
			ERPBranchID:         dept.ERPBranchID,
			IsCovered:           dept.IsCovered,
			BusinessID:          dept.BusinessID,
			BranchID:            dept.BranchID,
			CreatedAt:           dept.CreatedAt.Format(time.RFC3339),
			UpdatedAt:           dept.UpdatedAt.Format(time.RFC3339),
		})
	}

	c.JSON(http.StatusOK, utils.SuccessResponse{
		Status:  http.StatusOK,
		Message: "Departments retrieved successfully",
		Data: map[string]interface{}{
			"departments": departmentList,
			"pagination": map[string]interface{}{
				"total_records": paginationResult.TotalRecords,
				"total_pages":   paginationResult.TotalPages,
				"page":          paginationResult.Page,
				"per_page":      paginationResult.PerPage,
			},
		},
	})
}

func GetAllDepartmentsList(c *gin.Context, mongoClient *mongo.Client) {
	search := c.Query("search")

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
			{"department_short_name": bson.M{"$regex": search, "$options": "i"}},
		}
	}

	findOptions := options.Find().SetSort(bson.D{{Key: "name_en", Value: 1}})

	cursor, err := collections.Departments.Find(ctx, filter, findOptions)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: "Failed to retrieve departments",
		})
		return
	}
	defer cursor.Close(ctx)

	var departments []models.Department
	if err = cursor.All(ctx, &departments); err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: "Failed to decode departments",
		})
		return
	}

	departmentList := make([]dtos.DepartmentRes, 0, len(departments))
	for _, dept := range departments {
		departmentList = append(departmentList, dtos.DepartmentRes{
			ID:                  dept.ID.Hex(),
			NameEn:              dept.NameEn,
			NameAr:              dept.NameAr,
			DescriptionEn:       dept.DescriptionEn,
			DescriptionAr:       dept.DescriptionAr,
			DepartmentShortName: dept.DepartmentShortName,
			ERPBranchID:         dept.ERPBranchID,
			IsCovered:           dept.IsCovered,
			BusinessID:          dept.BusinessID,
			BranchID:            dept.BranchID,
			CreatedAt:           dept.CreatedAt.Format(time.RFC3339),
			UpdatedAt:           dept.UpdatedAt.Format(time.RFC3339),
		})
	}

	c.JSON(http.StatusOK, utils.SuccessResponse{
		Status:  http.StatusOK,
		Message: "Departments list retrieved successfully",
		Data: map[string]interface{}{
			"departments": departmentList,
			"total":       len(departmentList),
		},
	})
}

func GetDepartmentByID(c *gin.Context, mongoClient *mongo.Client) {
	departmentID := c.Param("id")

	objectID, err := primitive.ObjectIDFromHex(departmentID)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Status:  http.StatusBadRequest,
			Message: "Invalid department ID",
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

	var department models.Department
	filter := bson.M{
		"_id":         objectID,
		"business_id": user.BusinessId,
		"branch_id":   user.BranchId,
		"deleted_at":  bson.M{"$exists": false},
	}

	err = collections.Departments.FindOne(ctx, filter).Decode(&department)
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
			Message: "Failed to retrieve department",
		})
		return
	}

	departmentRes := dtos.DepartmentRes{
		ID:                  department.ID.Hex(),
		NameEn:              department.NameEn,
		NameAr:              department.NameAr,
		DescriptionEn:       department.DescriptionEn,
		DescriptionAr:       department.DescriptionAr,
		DepartmentShortName: department.DepartmentShortName,
		ERPBranchID:         department.ERPBranchID,
		IsCovered:           department.IsCovered,
		BusinessID:          department.BusinessID,
		BranchID:            department.BranchID,
		CreatedAt:           department.CreatedAt.Format(time.RFC3339),
		UpdatedAt:           department.UpdatedAt.Format(time.RFC3339),
	}

	c.JSON(http.StatusOK, utils.SuccessResponse{
		Status:  http.StatusOK,
		Message: "Department retrieved successfully",
		Data: map[string]interface{}{
			"department": departmentRes,
		},
	})
}

func UpdateDepartment(c *gin.Context, mongoClient *mongo.Client) {
	departmentID := c.Param("id")

	objectID, err := primitive.ObjectIDFromHex(departmentID)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Status:  http.StatusBadRequest,
			Message: "Invalid department ID",
		})
		return
	}

	var req dtos.UpdateDepartmentReq
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

	
	if req.NameEn != "" || req.NameAr != "" {
		existingFilter := bson.M{
			"business_id": user.BusinessId,
			"branch_id":   user.BranchId,
			"deleted_at":  bson.M{"$exists": false},
			"_id":         bson.M{"$ne": objectID},
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

		var existingDept models.Department
		err := collections.Departments.FindOne(ctx, existingFilter).Decode(&existingDept)
		if err == nil {
			c.JSON(http.StatusConflict, utils.ErrorResponse{
				Status:  http.StatusConflict,
				Message: "Department name already exists",
			})
			return
		} else if err != mongo.ErrNoDocuments {
			c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
				Status:  http.StatusInternalServerError,
				Message: "Failed to check department uniqueness: " + err.Error(),
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
	if req.DescriptionEn != "" {
		update["$set"].(bson.M)["description_en"] = req.DescriptionEn
	}
	if req.DescriptionAr != "" {
		update["$set"].(bson.M)["description_ar"] = req.DescriptionAr
	}
	if req.DepartmentShortName != "" {
		update["$set"].(bson.M)["department_short_name"] = req.DepartmentShortName
	}
	if req.ERPBranchID != 0 {
		update["$set"].(bson.M)["erp_branch_id"] = req.ERPBranchID
	}
	if req.IsCovered != nil {
		update["$set"].(bson.M)["is_covered"] = *req.IsCovered
	}

	result, err := collections.Departments.UpdateOne(ctx, filter, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: "Failed to update department: " + err.Error(),
		})
		return
	}

	if result.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, utils.ErrorResponse{
			Status:  http.StatusNotFound,
			Message: "Department not found",
		})
		return
	}

	var updatedDepartment models.Department
	err = collections.Departments.FindOne(ctx, filter).Decode(&updatedDepartment)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: "Failed to retrieve updated department",
		})
		return
	}

	departmentRes := dtos.DepartmentRes{
		ID:                  updatedDepartment.ID.Hex(),
		NameEn:              updatedDepartment.NameEn,
		NameAr:              updatedDepartment.NameAr,
		DescriptionEn:       updatedDepartment.DescriptionEn,
		DescriptionAr:       updatedDepartment.DescriptionAr,
		DepartmentShortName: updatedDepartment.DepartmentShortName,
		ERPBranchID:         updatedDepartment.ERPBranchID,
		IsCovered:           updatedDepartment.IsCovered,
		BusinessID:          updatedDepartment.BusinessID,
		BranchID:            updatedDepartment.BranchID,
		CreatedAt:           updatedDepartment.CreatedAt.Format(time.RFC3339),
		UpdatedAt:           updatedDepartment.UpdatedAt.Format(time.RFC3339),
	}

	c.JSON(http.StatusOK, utils.SuccessResponse{
		Status:  http.StatusOK,
		Message: "Department updated successfully",
		Data: map[string]interface{}{
			"department": departmentRes,
		},
	})
}

func DeleteDepartment(c *gin.Context, mongoClient *mongo.Client) {
	departmentID := c.Param("id")

	objectID, err := primitive.ObjectIDFromHex(departmentID)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Status:  http.StatusBadRequest,
			Message: "Invalid department ID",
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

	result, err := collections.Departments.UpdateOne(ctx, filter, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: "Failed to delete department: " + err.Error(),
		})
		return
	}

	if result.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, utils.ErrorResponse{
			Status:  http.StatusNotFound,
			Message: "Department not found",
		})
		return
	}

	c.JSON(http.StatusOK, utils.SuccessResponse{
		Status:  http.StatusOK,
		Message: "Department deleted successfully",
		Data:    nil,
	})
}

func SeedDepartments(c *gin.Context, mongoClient *mongo.Client) {
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

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	predefinedDepartments := []struct {
		NameEn              string
		NameAr              string
		DepartmentShortName string
		IsCovered           bool
	}{
		{NameEn: "Allergist", NameAr: "الحساسية", DepartmentShortName: "AL", IsCovered: false},
		{NameEn: "Anesthesiologist", NameAr: "التخدير", DepartmentShortName: "AN", IsCovered: false},
		{NameEn: "Cardiologist", NameAr: "أمراض القلب", DepartmentShortName: "CD", IsCovered: false},
		{NameEn: "Dermatologist", NameAr: "الأمراض الجلدية", DepartmentShortName: "DE", IsCovered: false},
		{NameEn: "Cosmetologist", NameAr: "التجميل", DepartmentShortName: "CO", IsCovered: false},
		{NameEn: "Endocrinologist", NameAr: "الغدد الصماء", DepartmentShortName: "EN", IsCovered: false},
		{NameEn: "Pediatric Endocrinologist", NameAr: "غدد صماء أطفال", DepartmentShortName: "PE", IsCovered: false},
		{NameEn: "Family Medicine", NameAr: "طب الأسرة", DepartmentShortName: "FM", IsCovered: false},
		{NameEn: "Gastroenterologist", NameAr: "الجهاز الهضمي", DepartmentShortName: "GE", IsCovered: false},
		{NameEn: "General Practitioner (GP)", NameAr: "طبيب عام", DepartmentShortName: "GP", IsCovered: false},
		{NameEn: "General Surgeon", NameAr: "الجراحة العامة", DepartmentShortName: "GS", IsCovered: false},
		{NameEn: "Plastic Surgeon", NameAr: "جراحة التجميل", DepartmentShortName: "PC", IsCovered: false},
		{NameEn: "Hematologist", NameAr: "أمراض الدم", DepartmentShortName: "HM", IsCovered: false},
		{NameEn: "Infectious Disease Specialist", NameAr: "الأمراض المعدية", DepartmentShortName: "ID", IsCovered: false},
		{NameEn: "Internal Medicine", NameAr: "الباطنة", DepartmentShortName: "IM", IsCovered: false},
		{NameEn: "Nephrologist", NameAr: "أمراض الكلى", DepartmentShortName: "NP", IsCovered: false},
		{NameEn: "Neurosurgeon", NameAr: "جراحة المخ والأعصاب", DepartmentShortName: "NS", IsCovered: false},
		{NameEn: "Obstetrician and Gynecologist", NameAr: "النساء والتوليد", DepartmentShortName: "OG", IsCovered: false},
		{NameEn: "Occupational Medicine Specialist", NameAr: "طب العمل", DepartmentShortName: "OM", IsCovered: false},
		{NameEn: "Oncologist", NameAr: "الأورام", DepartmentShortName: "ON", IsCovered: false},
		{NameEn: "Ophthalmologist", NameAr: "طب العيون", DepartmentShortName: "OP", IsCovered: false},
		{NameEn: "Orthopedic Surgeon", NameAr: "جراحة العظام", DepartmentShortName: "OR", IsCovered: false},
		{NameEn: "Otolaryngologist (ENT)", NameAr: "أنف وأذن وحنجرة", DepartmentShortName: "ET", IsCovered: false},
		{NameEn: "Pain Medicine Specialist", NameAr: "طب الألم", DepartmentShortName: "PM", IsCovered: false},
		{NameEn: "Pathologist", NameAr: "علم الأمراض", DepartmentShortName: "PA", IsCovered: false},
		{NameEn: "Pediatrician", NameAr: "طب الأطفال", DepartmentShortName: "PD", IsCovered: false},
		{NameEn: "Physiotherapist", NameAr: "العلاج الطبيعي", DepartmentShortName: "PH", IsCovered: false},
		{NameEn: "Psychiatrist", NameAr: "الطب النفسي", DepartmentShortName: "PY", IsCovered: false},
		{NameEn: "Psychologist", NameAr: "علم النفس", DepartmentShortName: "PG", IsCovered: false},
		{NameEn: "Pulmonologist", NameAr: "أمراض الصدر", DepartmentShortName: "PL", IsCovered: false},
		{NameEn: "Radiologist", NameAr: "الأشعة", DepartmentShortName: "RD", IsCovered: false},
		{NameEn: "Rheumatologist", NameAr: "الروماتيزم", DepartmentShortName: "RH", IsCovered: false},
		{NameEn: "Urologist", NameAr: "المسالك البولية", DepartmentShortName: "UR", IsCovered: false},
		{NameEn: "Vascular Surgeon", NameAr: "جراحة الأوعية الدموية", DepartmentShortName: "VS", IsCovered: false},
		{NameEn: "Bariatric Surgeon", NameAr: "جراحة السمنة", DepartmentShortName: "BS", IsCovered: false},
		{NameEn: "Breast Surgeon", NameAr: "جراحة الثدي", DepartmentShortName: "BR", IsCovered: false},
		{NameEn: "Cardiothoracic Surgeon", NameAr: "جراحة القلب والصدر", DepartmentShortName: "CT", IsCovered: false},
		{NameEn: "Colorectal Surgeon", NameAr: "جراحة القولون والمستقيم", DepartmentShortName: "CR", IsCovered: false},
		{NameEn: "Oral and Maxillofacial Surgeon", NameAr: "جراحة الفم والوجه والفكين", DepartmentShortName: "OF", IsCovered: false},
		{NameEn: "Pediatric Surgeon", NameAr: "جراحة الأطفال", DepartmentShortName: "PS", IsCovered: false},
		{NameEn: "Thoracic Surgeon", NameAr: "جراحة الصدر", DepartmentShortName: "TS", IsCovered: false},
		{NameEn: "Dentist", NameAr: "طب الأسنان", DepartmentShortName: "DN", IsCovered: false},
		{NameEn: "Orthodontics", NameAr: "تقويم الأسنان", DepartmentShortName: "OD", IsCovered: false},
		{NameEn: "Pediatric dentistry", NameAr: "أسنان الأطفال", DepartmentShortName: "DD", IsCovered: false},
		{NameEn: "Endodontics", NameAr: "علاج جذور الأسنان", DepartmentShortName: "ED", IsCovered: false},
		{NameEn: "Periodontics", NameAr: "أمراض اللثة", DepartmentShortName: "PR", IsCovered: false},
		{NameEn: "General Medicine", NameAr: "الطب العام", DepartmentShortName: "GM", IsCovered: false},
	}

	createdCount := 0
	skippedCount := 0
	var errors []string

	for _, dept := range predefinedDepartments {
		filter := bson.M{
			"name_en":     dept.NameEn,
			"business_id": user.BusinessId,
			"branch_id":   user.BranchId,
			"deleted_at":  bson.M{"$exists": false},
		}

		var existing models.Department
		err := collections.Departments.FindOne(ctx, filter).Decode(&existing)

		if err == mongo.ErrNoDocuments {
			department := models.Department{
				ID:                  primitive.NewObjectID(),
				NameEn:              dept.NameEn,
				NameAr:              dept.NameAr,
				DepartmentShortName: dept.DepartmentShortName,
				IsCovered:           dept.IsCovered,
				BusinessID:          user.BusinessId,
				BranchID:            user.BranchId,
				CreatedAt:           time.Now(),
				UpdatedAt:           time.Now(),
			}

			_, insertErr := collections.Departments.InsertOne(ctx, &department)
			if insertErr != nil {
				errors = append(errors, "Failed to create "+dept.NameEn+": "+insertErr.Error())
			} else {
				createdCount++
			}
		} else if err != nil {
			errors = append(errors, "Error checking "+dept.NameEn+": "+err.Error())
		} else {
			skippedCount++
		}
	}

	response := map[string]interface{}{
		"message":       "Departments seeded successfully",
		"created_count": createdCount,
		"skipped_count": skippedCount,
		"total_count":   len(predefinedDepartments),
		"business_id":   user.BusinessId,
		"branch_id":     user.BranchId,
	}

	if len(errors) > 0 {
		response["errors"] = errors
	}

	c.JSON(http.StatusOK, utils.SuccessResponse{
		Status:  http.StatusOK,
		Message: "Departments seeded successfully",
		Data:    response,
	})
}
