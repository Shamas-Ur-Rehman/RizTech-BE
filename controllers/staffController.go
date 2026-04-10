package controllers

import (
	"context"
	"fmt"
	"net/http"
	"supergit/inpatient/config"
	"supergit/inpatient/dtos"
	"supergit/inpatient/models"
	"supergit/inpatient/utils"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"
)

func CreateStaff(c *gin.Context, sqlDB *gorm.DB, mongoClient *mongo.Client) {
	var req dtos.CreateStaffReq

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Status:  http.StatusBadRequest,
			Message: "Invalid request: " + err.Error(),
		})
		return
	}

	user, exists := utils.GetBusinessAndBranch(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, utils.ErrorResponse{
			Status:  http.StatusUnauthorized,
			Message: "User not authenticated",
		})
		return
	}

	if req.UserData == nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Status:  http.StatusBadRequest,
			Message: "user_data is required",
		})
		return
	}

	var existingUser models.User
	if err := sqlDB.Where("email = ?", req.UserData.Email).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, utils.ErrorResponse{
			Status:  http.StatusConflict,
			Message: "Email already exists",
		})
		return
	}

	hashedPassword, err := utils.HashPassword(req.UserData.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: "Failed to hash password",
		})
		return
	}

	employeeId, err := utils.GenerateNextEmployeeID(sqlDB)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: "Failed to generate employee ID",
		})
		return
	}

	newUser := models.User{
		FullName:    req.UserData.FullName,
		Email:       req.UserData.Email,
		Password:    hashedPassword,
		EmployeeId:  employeeId,
		Contact:     req.UserData.Contact,
		Gender:      req.UserData.Gender,
		RoleID:      req.UserData.RoleID,
		Service:     "inpatient",
		Address:     req.UserData.Address,
		Nationality: req.UserData.Nationality,
		DocumentId:  req.UserData.DocumentId,
		License:     req.UserData.License,
		BusinessId:  user.BusinessID,
		BranchId:    user.BranchID,
		IsActive:    true,
		IsStaff:     true,
	}

	if req.UserData.DateOfBirth != nil && *req.UserData.DateOfBirth != "" {
		dob, err := time.Parse("2006-01-02", *req.UserData.DateOfBirth)
		if err != nil {
			c.JSON(http.StatusBadRequest, utils.ErrorResponse{
				Status:  http.StatusBadRequest,
				Message: "Invalid date of birth format. Use YYYY-MM-DD",
			})
			return
		}
		newUser.DateOfBirth = &dob
	}

	if err := sqlDB.Create(&newUser).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: "Failed to create user: " + err.Error(),
		})
		return
	}

	userID := newUser.ID

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	staffCollection := config.GetCollections(mongoClient, "inpatient").Staff

	var existingStaff models.Staff
	err = staffCollection.FindOne(ctx, bson.M{
		"user_id":     userID,
		"business_id": user.BusinessID,
		"branch_id":   user.BranchID,
		"deleted_at":  bson.M{"$exists": false},
	}).Decode(&existingStaff)

	if err == nil {
		c.JSON(http.StatusConflict, utils.ErrorResponse{
			Status:  http.StatusConflict,
			Message: "Staff record already exists for this user",
		})
		return
	}

	schedules := make([]models.Schedule, 0, len(req.Schedule))
	for _, schedReq := range req.Schedule {
		shiftIDs := make([]primitive.ObjectID, 0, len(schedReq.Shifts))
		for _, shiftIDStr := range schedReq.Shifts {
			shiftID, err := primitive.ObjectIDFromHex(shiftIDStr)
			if err != nil {
				c.JSON(http.StatusBadRequest, utils.ErrorResponse{
					Status:  http.StatusBadRequest,
					Message: fmt.Sprintf("Invalid shift ID format: %s", shiftIDStr),
				})
				return
			}
			shiftIDs = append(shiftIDs, shiftID)
		}
		schedules = append(schedules, models.Schedule{
			Day:    schedReq.Day,
			Shifts: shiftIDs,
		})
	}

	staff := models.Staff{
		ID:         primitive.NewObjectID(),
		UserID:     userID,
		BusinessID: user.BusinessID,
		BranchID:   user.BranchID,
		Schedule:   schedules,
		WardNo:     req.WardNo,
		RoomNo:     req.RoomNo,
		Available:  req.Available,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if req.DepartmentID != "" {
		deptID, err := primitive.ObjectIDFromHex(req.DepartmentID)
		if err != nil {
			c.JSON(http.StatusBadRequest, utils.ErrorResponse{
				Status:  http.StatusBadRequest,
				Message: "Invalid department ID format",
			})
			return
		}
		staff.DepartmentID = deptID
	}

	if req.SpecialityID != "" {
		specID, err := primitive.ObjectIDFromHex(req.SpecialityID)
		if err != nil {
			c.JSON(http.StatusBadRequest, utils.ErrorResponse{
				Status:  http.StatusBadRequest,
				Message: "Invalid speciality ID format",
			})
			return
		}
		staff.SpecialityID = specID
	}

	if len(req.AssignedNurses) > 0 {
		assignedNurses := make([]primitive.ObjectID, 0, len(req.AssignedNurses))
		for _, nurseIDStr := range req.AssignedNurses {
			nurseID, err := primitive.ObjectIDFromHex(nurseIDStr)
			if err != nil {
				c.JSON(http.StatusBadRequest, utils.ErrorResponse{
					Status:  http.StatusBadRequest,
					Message: fmt.Sprintf("Invalid nurse ID format: %s", nurseIDStr),
				})
				return
			}
			assignedNurses = append(assignedNurses, nurseID)
		}
		staff.AssignedNurses = assignedNurses
	}

	_, err = staffCollection.InsertOne(ctx, staff)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: "Failed to create staff: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, utils.SuccessResponse{
		Status:  http.StatusCreated,
		Message: "Staff created successfully",
		Data: map[string]interface{}{
			"staff": dtos.StaffRes{
				ID:             staff.ID,
				UserID:         staff.UserID,
				BusinessID:     staff.BusinessID,
				BranchID:       staff.BranchID,
				DepartmentID:   staff.DepartmentID,
				SpecialityID:   staff.SpecialityID,
				Schedule:       staff.Schedule,
				AssignedNurses: staff.AssignedNurses,
				WardNo:         staff.WardNo,
				RoomNo:         staff.RoomNo,
				Available:      staff.Available,
				CreatedAt:      staff.CreatedAt,
				UpdatedAt:      staff.UpdatedAt,
			},
		},
	})
}

func GetAllStaff(c *gin.Context, sqlDB *gorm.DB, mongoClient *mongo.Client) {
	user, exists := utils.GetBusinessAndBranch(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, utils.ErrorResponse{
			Status:  http.StatusUnauthorized,
			Message: "User not authenticated",
		})
		return
	}

	page := utils.StringToInt(c.DefaultQuery("page", "1"))
	perPage := utils.StringToInt(c.DefaultQuery("per_page", "10"))
	search := c.Query("search")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	staffCollection := config.GetCollections(mongoClient, "inpatient").Staff

	filter := bson.M{
		"business_id": user.BusinessID,
		"branch_id":   user.BranchID,
		"deleted_at":  bson.M{"$exists": false},
	}

	if search != "" {
		filter["staff_id"] = bson.M{"$regex": search, "$options": "i"}
	}

	var staffList []models.Staff
	result, err := utils.PaginateMongo(ctx, staffCollection, filter, page, perPage, &staffList, "created_at", -1)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: "Failed to retrieve staff: " + err.Error(),
		})
		return
	}

	userIDs := make([]uint, 0, len(staffList))
	for _, staff := range staffList {
		userIDs = append(userIDs, staff.UserID)
	}

	var users []models.User
	if len(userIDs) > 0 {
		sqlDB.Where("id IN ?", userIDs).Find(&users)
	}

	userMap := make(map[uint]models.User)
	for _, user := range users {
		userMap[user.ID] = user
	}

	staffResList := make([]dtos.StaffRes, 0, len(staffList))
	for _, staff := range staffList {
		user := userMap[staff.UserID]
		staffRes := dtos.StaffRes{
			ID:             staff.ID,
			UserID:         staff.UserID,
			BusinessID:     staff.BusinessID,
			BranchID:       staff.BranchID,
			DepartmentID:   staff.DepartmentID,
			SpecialityID:   staff.SpecialityID,
			Schedule:       staff.Schedule,
			AssignedNurses: staff.AssignedNurses,
			WardNo:         staff.WardNo,
			RoomNo:         staff.RoomNo,
			Available:      staff.Available,
			CreatedAt:      staff.CreatedAt,
			UpdatedAt:      staff.UpdatedAt,
		}

		if user.ID != 0 {
			staffRes.User = &dtos.UserRes{
				ID:       user.ID,
				FullName: user.FullName,
				Email:    user.Email,
				Contact:  user.Contact,
				Gender:   user.Gender,
			}
		}

		staffResList = append(staffResList, staffRes)
	}

	c.JSON(http.StatusOK, utils.SuccessResponse{
		Status:  http.StatusOK,
		Message: "Staff retrieved successfully",
		Data: map[string]interface{}{
			"staff": staffResList,
			"pagination": map[string]interface{}{
				"total_records": result.TotalRecords,
				"total_pages":   result.TotalPages,
				"page":          result.Page,
				"per_page":      result.PerPage,
			},
		},
	})
}

func GetAllStaffList(c *gin.Context, sqlDB *gorm.DB, mongoClient *mongo.Client) {
	user, exists := utils.GetBusinessAndBranch(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, utils.ErrorResponse{
			Status:  http.StatusUnauthorized,
			Message: "User not authenticated",
		})
		return
	}

	roleName := c.Query("role")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	staffCollection := config.GetCollections(mongoClient, "inpatient").Staff

	filter := bson.M{
		"business_id": user.BusinessID,
		"branch_id":   user.BranchID,
		"deleted_at":  bson.M{"$exists": false},
	}

	var userIDs []uint
	if roleName != "" {
		var role models.Role
		if err := sqlDB.Where("LOWER(name) = LOWER(?) AND business_id = ? AND branch_id = ? AND deleted_at IS NULL",
			roleName, user.BusinessID, user.BranchID).First(&role).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, utils.ErrorResponse{
					Status:  http.StatusNotFound,
					Message: "Role not found",
				})
				return
			}
			c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
				Status:  http.StatusInternalServerError,
				Message: "Failed to retrieve role: " + err.Error(),
			})
			return
		}

		var users []models.User
		if err := sqlDB.Where("role_id = ? AND business_id = ? AND branch_id = ? AND is_staff = ? AND is_active = ? AND deleted_at IS NULL",
			role.ID, user.BusinessID, user.BranchID, true, true).Find(&users).Error; err != nil {
			c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
				Status:  http.StatusInternalServerError,
				Message: "Failed to retrieve users: " + err.Error(),
			})
			return
		}

		if len(users) == 0 {
			c.JSON(http.StatusOK, utils.SuccessResponse{
				Status:  http.StatusOK,
				Message: "No staff found with this role",
				Data: map[string]interface{}{
					"staff": []interface{}{},
					"total": 0,
				},
			})
			return
		}

		userIDs = make([]uint, len(users))
		for i, u := range users {
			userIDs[i] = u.ID
		}

		filter["user_id"] = bson.M{"$in": userIDs}
	}

	cursor, err := staffCollection.Find(ctx, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: "Failed to retrieve staff",
		})
		return
	}
	defer cursor.Close(ctx)

	var staffList []models.Staff
	if err = cursor.All(ctx, &staffList); err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: "Failed to decode staff",
		})
		return
	}

	if roleName != "" {
		staffUserIDs := make([]uint, len(staffList))
		for i, staff := range staffList {
			staffUserIDs[i] = staff.UserID
		}

		var users []models.User
		if len(staffUserIDs) > 0 {
			sqlDB.Where("id IN ?", staffUserIDs).Preload("Role").Find(&users)
		}

		userMap := make(map[uint]models.User)
		for _, u := range users {
			userMap[u.ID] = u
		}

		staffResList := make([]dtos.StaffRes, 0, len(staffList))
		for _, staff := range staffList {
			u := userMap[staff.UserID]
			staffRes := dtos.StaffRes{
				ID:             staff.ID,
				UserID:         staff.UserID,
				BusinessID:     staff.BusinessID,
				BranchID:       staff.BranchID,
				DepartmentID:   staff.DepartmentID,
				SpecialityID:   staff.SpecialityID,
				Schedule:       staff.Schedule,
				AssignedNurses: staff.AssignedNurses,
				WardNo:         staff.WardNo,
				RoomNo:         staff.RoomNo,
				Available:      staff.Available,
				CreatedAt:      staff.CreatedAt,
				UpdatedAt:      staff.UpdatedAt,
			}

			if u.ID != 0 {
				staffRes.User = &dtos.UserRes{
					ID:         u.ID,
					FullName:   u.FullName,
					Email:      u.Email,
					Contact:    u.Contact,
					Gender:     u.Gender,
					EmployeeId: u.EmployeeId,
					RoleID:     u.RoleID,
					RoleName:   u.Role.Name,
					IsActive:   u.IsActive,
					IsStaff:    u.IsStaff,
				}

				if u.DateOfBirth != nil {
					dobStr := u.DateOfBirth.Format("2006-01-02")
					staffRes.User.DateOfBirth = &dobStr
					age := utils.CalculateAge(*u.DateOfBirth)
					staffRes.User.Age = &age
				}
			}

			staffResList = append(staffResList, staffRes)
		}

		c.JSON(http.StatusOK, utils.SuccessResponse{
			Status:  http.StatusOK,
			Message: "Staff list retrieved successfully",
			Data: map[string]interface{}{
				"staff": staffResList,
				"total": len(staffResList),
			},
		})
		return
	}

	c.JSON(http.StatusOK, utils.SuccessResponse{
		Status:  http.StatusOK,
		Message: "Staff list retrieved successfully",
		Data: map[string]interface{}{
			"staff": staffList,
			"total": len(staffList),
		},
	})
}



func GetStaffByID(c *gin.Context, sqlDB *gorm.DB, mongoClient *mongo.Client) {
	staffID := c.Param("id")
	user, exists := utils.GetBusinessAndBranch(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, utils.ErrorResponse{
			Status:  http.StatusUnauthorized,
			Message: "User not authenticated",
		})
		return
	}

	objectID, err := primitive.ObjectIDFromHex(staffID)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Status:  http.StatusBadRequest,
			Message: "Invalid staff ID format",
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	staffCollection := config.GetCollections(mongoClient, "inpatient").Staff

	var staff models.Staff
	err = staffCollection.FindOne(ctx, bson.M{
		"_id":         objectID,
		"business_id": user.BusinessID,
		"branch_id":   user.BranchID,
		"deleted_at":  bson.M{"$exists": false},
	}).Decode(&staff)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, utils.ErrorResponse{
				Status:  http.StatusNotFound,
				Message: "Staff not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: "Failed to retrieve staff",
		})
		return
	}

	var dbUser models.User
	sqlDB.First(&dbUser, staff.UserID)

	staffRes := dtos.StaffRes{
		ID:             staff.ID,
		UserID:         staff.UserID,
		BusinessID:     staff.BusinessID,
		BranchID:       staff.BranchID,
		DepartmentID:   staff.DepartmentID,
		SpecialityID:   staff.SpecialityID,
		Schedule:       staff.Schedule,
		AssignedNurses: staff.AssignedNurses,
		WardNo:         staff.WardNo,
		RoomNo:         staff.RoomNo,
		Available:      staff.Available,
		CreatedAt:      staff.CreatedAt,
		UpdatedAt:      staff.UpdatedAt,
	}

	if dbUser.ID != 0 {
		staffRes.User = &dtos.UserRes{
			ID:          dbUser.ID,
			FullName:    dbUser.FullName,
			Email:       dbUser.Email,
			Contact:     dbUser.Contact,
			Gender:      dbUser.Gender,
			EmployeeId:  dbUser.EmployeeId,
			DocumentId:  dbUser.DocumentId,
			License:     dbUser.License,
			Address:     dbUser.Address,
			Nationality: dbUser.Nationality,
		}
	}

	c.JSON(http.StatusOK, utils.SuccessResponse{
		Status:  http.StatusOK,
		Message: "Staff retrieved successfully",
		Data: map[string]interface{}{
			"staff": staffRes,
		},
	})
}

func UpdateStaff(c *gin.Context, sqlDB *gorm.DB, mongoClient *mongo.Client) {
	staffID := c.Param("id")
	user, exists := utils.GetBusinessAndBranch(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, utils.ErrorResponse{
			Status:  http.StatusUnauthorized,
			Message: "User not authenticated",
		})
		return
	}

	objectID, err := primitive.ObjectIDFromHex(staffID)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Status:  http.StatusBadRequest,
			Message: "Invalid staff ID format",
		})
		return
	}

	var req dtos.UpdateStaffReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Status:  http.StatusBadRequest,
			Message: "Invalid request: " + err.Error(),
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	staffCollection := config.GetCollections(mongoClient, "inpatient").Staff

	update := bson.M{
		"$set": bson.M{
			"updated_at": time.Now(),
		},
	}

	if req.DepartmentID != "" {
		deptID, err := primitive.ObjectIDFromHex(req.DepartmentID)
		if err != nil {
			c.JSON(http.StatusBadRequest, utils.ErrorResponse{
				Status:  http.StatusBadRequest,
				Message: "Invalid department ID format",
			})
			return
		}
		update["$set"].(bson.M)["department_id"] = deptID
	}

	if req.SpecialityID != "" {
		specID, err := primitive.ObjectIDFromHex(req.SpecialityID)
		if err != nil {
			c.JSON(http.StatusBadRequest, utils.ErrorResponse{
				Status:  http.StatusBadRequest,
				Message: "Invalid speciality ID format",
			})
			return
		}
		update["$set"].(bson.M)["speciality_id"] = specID
	}

	if req.Schedule != nil {
		schedules := make([]models.Schedule, 0, len(req.Schedule))
		for _, schedReq := range req.Schedule {
			shiftIDs := make([]primitive.ObjectID, 0, len(schedReq.Shifts))
			for _, shiftIDStr := range schedReq.Shifts {
				shiftID, err := primitive.ObjectIDFromHex(shiftIDStr)
				if err != nil {
					c.JSON(http.StatusBadRequest, utils.ErrorResponse{
						Status:  http.StatusBadRequest,
						Message: fmt.Sprintf("Invalid shift ID format: %s", shiftIDStr),
					})
					return
				}
				shiftIDs = append(shiftIDs, shiftID)
			}
			schedules = append(schedules, models.Schedule{
				Day:    schedReq.Day,
				Shifts: shiftIDs,
			})
		}
		update["$set"].(bson.M)["schedule"] = schedules
	}

	if req.AssignedNurses != nil {
		assignedNurses := make([]primitive.ObjectID, 0, len(req.AssignedNurses))
		for _, nurseIDStr := range req.AssignedNurses {
			nurseID, err := primitive.ObjectIDFromHex(nurseIDStr)
			if err != nil {
				c.JSON(http.StatusBadRequest, utils.ErrorResponse{
					Status:  http.StatusBadRequest,
					Message: fmt.Sprintf("Invalid nurse ID format: %s", nurseIDStr),
				})
				return
			}
			assignedNurses = append(assignedNurses, nurseID)
		}
		update["$set"].(bson.M)["assigned_nurses"] = assignedNurses
	}

	if req.WardNo != "" {
		update["$set"].(bson.M)["ward_no"] = req.WardNo
	}

	if req.RoomNo != "" {
		update["$set"].(bson.M)["room_no"] = req.RoomNo
	}

	if req.Available != nil {
		update["$set"].(bson.M)["available"] = *req.Available
	}

	result, err := staffCollection.UpdateOne(ctx, bson.M{
		"_id":         objectID,
		"business_id": user.BusinessID,
		"branch_id":   user.BranchID,
		"deleted_at":  bson.M{"$exists": false},
	}, update)

	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: "Failed to update staff: " + err.Error(),
		})
		return
	}

	if result.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, utils.ErrorResponse{
			Status:  http.StatusNotFound,
			Message: "Staff not found",
		})
		return
	}

	var staff models.Staff
	err = staffCollection.FindOne(ctx, bson.M{
		"_id":         objectID,
		"business_id": user.BusinessID,
		"branch_id":   user.BranchID,
		"deleted_at":  bson.M{"$exists": false},
	}).Decode(&staff)

	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: "Failed to retrieve updated staff",
		})
		return
	}

	c.JSON(http.StatusOK, utils.SuccessResponse{
		Status:  http.StatusOK,
		Message: "Staff updated successfully",
		Data: map[string]interface{}{
			"staff": dtos.StaffRes{
				ID:             staff.ID,
				UserID:         staff.UserID,
				BusinessID:     staff.BusinessID,
				BranchID:       staff.BranchID,
				DepartmentID:   staff.DepartmentID,
				SpecialityID:   staff.SpecialityID,
				Schedule:       staff.Schedule,
				AssignedNurses: staff.AssignedNurses,
				WardNo:         staff.WardNo,
				RoomNo:         staff.RoomNo,
				Available:      staff.Available,
				CreatedAt:      staff.CreatedAt,
				UpdatedAt:      staff.UpdatedAt,
			},
		},
	})
}

func DeleteStaff(c *gin.Context, mongoClient *mongo.Client) {
	staffID := c.Param("id")
	user, exists := utils.GetBusinessAndBranch(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, utils.ErrorResponse{
			Status:  http.StatusUnauthorized,
			Message: "User not authenticated",
		})
		return
	}

	objectID, err := primitive.ObjectIDFromHex(staffID)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Status:  http.StatusBadRequest,
			Message: "Invalid staff ID format",
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	staffCollection := config.GetCollections(mongoClient, "inpatient").Staff

	result, err := staffCollection.UpdateOne(ctx, bson.M{
		"_id":         objectID,
		"business_id": user.BusinessID,
		"branch_id":   user.BranchID,
		"deleted_at":  bson.M{"$exists": false},
	}, bson.M{
		"$set": bson.M{
			"deleted_at": time.Now(),
		},
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: "Failed to delete staff: " + err.Error(),
		})
		return
	}

	if result.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, utils.ErrorResponse{
			Status:  http.StatusNotFound,
			Message: "Staff not found",
		})
		return
	}

	c.JSON(http.StatusOK, utils.SuccessResponse{
		Status:  http.StatusOK,
		Message: "Staff deleted successfully",
		Data:    nil,
	})
}
