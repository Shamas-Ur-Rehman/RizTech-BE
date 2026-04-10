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
)

func CreateShift(c *gin.Context, mongoClient *mongo.Client) {
	var req dtos.CreateShiftReq

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
		"title":       bson.M{"$regex": "^" + req.Title + "$", "$options": "i"},
	}

	var existingShift models.Shift
	err := collections.Shifts.FindOne(ctx, existingFilter).Decode(&existingShift)
	if err == nil {
		c.JSON(http.StatusConflict, utils.ErrorResponse{
			Status:  http.StatusConflict,
			Message: "Shift title already exists",
		})
		return
	} else if err != mongo.ErrNoDocuments {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: "Failed to check shift uniqueness: " + err.Error(),
		})
		return
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	shift := models.Shift{
		ID:         primitive.NewObjectID(),
		Title:      req.Title,
		StartTime:  req.StartTime,
		EndTime:    req.EndTime,
		IsActive:   isActive,
		BusinessID: user.BusinessId,
		BranchID:   user.BranchId,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	_, err = collections.Shifts.InsertOne(ctx, shift)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: "Failed to create shift: " + err.Error(),
		})
		return
	}

	shiftRes := dtos.ShiftRes{
		ID:         shift.ID.Hex(),
		Title:      shift.Title,
		StartTime:  shift.StartTime,
		EndTime:    shift.EndTime,
		IsActive:   shift.IsActive,
		BusinessID: shift.BusinessID,
		BranchID:   shift.BranchID,
		CreatedAt:  shift.CreatedAt.Format(time.RFC3339),
		UpdatedAt:  shift.UpdatedAt.Format(time.RFC3339),
	}

	c.JSON(http.StatusCreated, utils.SuccessResponse{
		Status:  http.StatusCreated,
		Message: "Shift created successfully",
		Data: map[string]interface{}{
			"shift": shiftRes,
		},
	})
}

func GetAllShifts(c *gin.Context, mongoClient *mongo.Client) {
	page := utils.StringToInt(c.DefaultQuery("page", "1"))
	perPage := utils.StringToInt(c.DefaultQuery("per_page", "10"))
	search := c.Query("search")
	isActive := c.Query("is_active")

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
		filter["title"] = bson.M{"$regex": search, "$options": "i"}
	}

	if isActive != "" {
		if isActive == "true" {
			filter["is_active"] = true
		} else if isActive == "false" {
			filter["is_active"] = false
		}
	}

	var shifts []models.Shift
	paginationResult, err := utils.PaginateMongo(ctx, collections.Shifts, filter, page, perPage, &shifts, "created_at", -1)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: "Failed to retrieve shifts: " + err.Error(),
		})
		return
	}

	shiftList := make([]dtos.ShiftRes, 0, len(shifts))
	for _, shift := range shifts {
		shiftList = append(shiftList, dtos.ShiftRes{
			ID:         shift.ID.Hex(),
			Title:      shift.Title,
			StartTime:  shift.StartTime,
			EndTime:    shift.EndTime,
			IsActive:   shift.IsActive,
			BusinessID: shift.BusinessID,
			BranchID:   shift.BranchID,
			CreatedAt:  shift.CreatedAt.Format(time.RFC3339),
			UpdatedAt:  shift.UpdatedAt.Format(time.RFC3339),
		})
	}

	c.JSON(http.StatusOK, utils.SuccessResponse{
		Status:  http.StatusOK,
		Message: "Shifts retrieved successfully",
		Data: map[string]interface{}{
			"shifts": shiftList,
			"pagination": map[string]interface{}{
				"total_records": paginationResult.TotalRecords,
				"total_pages":   paginationResult.TotalPages,
				"page":          paginationResult.Page,
				"per_page":      paginationResult.PerPage,
			},
		},
	})
}

func GetAllShiftsList(c *gin.Context, mongoClient *mongo.Client) {
	search := c.Query("search")
	isActive := c.Query("is_active")

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
		filter["title"] = bson.M{"$regex": search, "$options": "i"}
	}

	if isActive != "" {
		if isActive == "true" {
			filter["is_active"] = true
		} else if isActive == "false" {
			filter["is_active"] = false
		}
	}

	cursor, err := collections.Shifts.Find(ctx, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: "Failed to retrieve shifts",
		})
		return
	}
	defer cursor.Close(ctx)

	var shifts []models.Shift
	if err = cursor.All(ctx, &shifts); err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: "Failed to decode shifts",
		})
		return
	}

	shiftList := make([]dtos.ShiftRes, 0, len(shifts))
	for _, shift := range shifts {
		shiftList = append(shiftList, dtos.ShiftRes{
			ID:         shift.ID.Hex(),
			Title:      shift.Title,
			StartTime:  shift.StartTime,
			EndTime:    shift.EndTime,
			IsActive:   shift.IsActive,
			BusinessID: shift.BusinessID,
			BranchID:   shift.BranchID,
			CreatedAt:  shift.CreatedAt.Format(time.RFC3339),
			UpdatedAt:  shift.UpdatedAt.Format(time.RFC3339),
		})
	}

	c.JSON(http.StatusOK, utils.SuccessResponse{
		Status:  http.StatusOK,
		Message: "Shifts list retrieved successfully",
		Data: map[string]interface{}{
			"shifts": shiftList,
			"total":  len(shiftList),
		},
	})
}

func GetShiftByID(c *gin.Context, mongoClient *mongo.Client) {
	shiftID := c.Param("id")

	objectID, err := primitive.ObjectIDFromHex(shiftID)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Status:  http.StatusBadRequest,
			Message: "Invalid shift ID",
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

	var shift models.Shift
	filter := bson.M{
		"_id":         objectID,
		"business_id": user.BusinessId,
		"branch_id":   user.BranchId,
		"deleted_at":  bson.M{"$exists": false},
	}

	err = collections.Shifts.FindOne(ctx, filter).Decode(&shift)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, utils.ErrorResponse{
				Status:  http.StatusNotFound,
				Message: "Shift not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: "Failed to retrieve shift",
		})
		return
	}

	shiftRes := dtos.ShiftRes{
		ID:         shift.ID.Hex(),
		Title:      shift.Title,
		StartTime:  shift.StartTime,
		EndTime:    shift.EndTime,
		IsActive:   shift.IsActive,
		BusinessID: shift.BusinessID,
		BranchID:   shift.BranchID,
		CreatedAt:  shift.CreatedAt.Format(time.RFC3339),
		UpdatedAt:  shift.UpdatedAt.Format(time.RFC3339),
	}

	c.JSON(http.StatusOK, utils.SuccessResponse{
		Status:  http.StatusOK,
		Message: "Shift retrieved successfully",
		Data: map[string]interface{}{
			"shift": shiftRes,
		},
	})
}

func UpdateShift(c *gin.Context, mongoClient *mongo.Client) {
	shiftID := c.Param("id")

	objectID, err := primitive.ObjectIDFromHex(shiftID)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Status:  http.StatusBadRequest,
			Message: "Invalid shift ID",
		})
		return
	}

	var req dtos.UpdateShiftReq
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

	filter := bson.M{
		"_id":         objectID,
		"business_id": user.BusinessId,
		"branch_id":   user.BranchId,
		"deleted_at":  bson.M{"$exists": false},
	}

	if req.Title != "" {
		existingFilter := bson.M{
			"business_id": user.BusinessId,
			"branch_id":   user.BranchId,
			"deleted_at":  bson.M{"$exists": false},
			"title":       bson.M{"$regex": "^" + req.Title + "$", "$options": "i"},
			"_id":         bson.M{"$ne": objectID},
		}

		var existingShift models.Shift
		err := collections.Shifts.FindOne(ctx, existingFilter).Decode(&existingShift)
		if err == nil {
			c.JSON(http.StatusConflict, utils.ErrorResponse{
				Status:  http.StatusConflict,
				Message: "Shift title already exists",
			})
			return
		} else if err != mongo.ErrNoDocuments {
			c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
				Status:  http.StatusInternalServerError,
				Message: "Failed to check shift uniqueness: " + err.Error(),
			})
			return
		}
	}

	update := bson.M{
		"$set": bson.M{
			"updated_at": time.Now(),
		},
	}

	if req.Title != "" {
		update["$set"].(bson.M)["title"] = req.Title
	}
	if req.StartTime != "" {
		update["$set"].(bson.M)["start_time"] = req.StartTime
	}
	if req.EndTime != "" {
		update["$set"].(bson.M)["end_time"] = req.EndTime
	}
	if req.IsActive != nil {
		update["$set"].(bson.M)["is_active"] = *req.IsActive
	}

	result, err := collections.Shifts.UpdateOne(ctx, filter, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: "Failed to update shift: " + err.Error(),
		})
		return
	}

	if result.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, utils.ErrorResponse{
			Status:  http.StatusNotFound,
			Message: "Shift not found",
		})
		return
	}

	var updatedShift models.Shift
	err = collections.Shifts.FindOne(ctx, filter).Decode(&updatedShift)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: "Failed to retrieve updated shift",
		})
		return
	}

	shiftRes := dtos.ShiftRes{
		ID:         updatedShift.ID.Hex(),
		Title:      updatedShift.Title,
		StartTime:  updatedShift.StartTime,
		EndTime:    updatedShift.EndTime,
		IsActive:   updatedShift.IsActive,
		BusinessID: updatedShift.BusinessID,
		BranchID:   updatedShift.BranchID,
		CreatedAt:  updatedShift.CreatedAt.Format(time.RFC3339),
		UpdatedAt:  updatedShift.UpdatedAt.Format(time.RFC3339),
	}

	c.JSON(http.StatusOK, utils.SuccessResponse{
		Status:  http.StatusOK,
		Message: "Shift updated successfully",
		Data: map[string]interface{}{
			"shift": shiftRes,
		},
	})
}

func DeleteShift(c *gin.Context, mongoClient *mongo.Client) {
	shiftID := c.Param("id")

	objectID, err := primitive.ObjectIDFromHex(shiftID)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Status:  http.StatusBadRequest,
			Message: "Invalid shift ID",
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

	result, err := collections.Shifts.UpdateOne(ctx, filter, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: "Failed to delete shift: " + err.Error(),
		})
		return
	}

	if result.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, utils.ErrorResponse{
			Status:  http.StatusNotFound,
			Message: "Shift not found",
		})
		return
	}

	c.JSON(http.StatusOK, utils.SuccessResponse{
		Status:  http.StatusOK,
		Message: "Shift deleted successfully",
		Data:    nil,
	})
}
