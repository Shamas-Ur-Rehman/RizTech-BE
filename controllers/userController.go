package controllers

import (
	"net/http"
	"supergit/inpatient/dtos"
	"supergit/inpatient/models"
	"supergit/inpatient/utils"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func CreateUser(c *gin.Context, sqlDB *gorm.DB) {
	var req struct {
		FullName    string  `json:"full_name" binding:"required"`
		Email       string  `json:"email" binding:"required,email"`
		Password    string  `json:"password" binding:"required,min=8"`
		EmployeeId  string  `json:"employee_id"`
		Contact     string  `json:"contact"`
		DateOfBirth *string `json:"date_of_birth"` // Format: "2006-01-02"
		Gender      string  `json:"gender"`        // male, female, other
		RoleID      uint    `json:"role_id" binding:"required"`
		Address     string  `json:"address"`
		Nationality string  `json:"nationality"`
		DocumentId  string  `json:"document_id"`
		License     string  `json:"license"`
		IsStaff     bool    `json:"is_staff"`
		BusinessId  uint    `json:"business_id" binding:"required"`
		BranchId    uint    `json:"branch_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Status:  http.StatusBadRequest,
			Message: "Invalid request: " + err.Error(),
		})
		return
	}
	var existingUser models.User
	if err := sqlDB.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, utils.ErrorResponse{
			Status:  http.StatusConflict,
			Message: "Email already exists",
		})
		return
	}
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: "Failed to hash password",
		})
		return
	}
	employeeId := req.EmployeeId
	if employeeId == "" {
		generatedId, err := utils.GenerateNextEmployeeID(sqlDB)
		if err != nil {
			c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
				Status:  http.StatusInternalServerError,
				Message: "Failed to generate employee ID",
			})
			return
		}
		employeeId = generatedId
	} else {
		var existingEmp models.User
		if err := sqlDB.Where("employee_id = ?", employeeId).First(&existingEmp).Error; err == nil {
			c.JSON(http.StatusConflict, utils.ErrorResponse{
				Status:  http.StatusConflict,
				Message: "Employee ID already exists",
			})
			return
		}
	}
	user := models.User{
		FullName:    req.FullName,
		Email:       req.Email,
		Password:    hashedPassword,
		EmployeeId:  employeeId,
		Contact:     req.Contact,
		Gender:      req.Gender,
		RoleID:      req.RoleID,
		Service:     "inpatient",
		Address:     req.Address,
		Nationality: req.Nationality,
		DocumentId:  req.DocumentId,
		License:     req.License,
		BusinessId:  req.BusinessId,
		BranchId:    req.BranchId,
		IsActive:    true,
		IsStaff:     req.IsStaff,
	}

	if req.DateOfBirth != nil && *req.DateOfBirth != "" {
		dob, err := time.Parse("2006-01-02", *req.DateOfBirth)
		if err != nil {
			c.JSON(http.StatusBadRequest, utils.ErrorResponse{
				Status:  http.StatusBadRequest,
				Message: "Invalid date of birth format. Use YYYY-MM-DD",
			})
			return
		}
		user.DateOfBirth = &dob
	}

	if err := sqlDB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: "Failed to create user: " + err.Error(),
		})

		return
	}
	var role models.Role
	var business models.Business
	sqlDB.Select("id, name").First(&role, user.RoleID)
	sqlDB.Select("id, name_en, name_ar").First(&business, user.BusinessId)

	user.Role = role
	user.Business = business

	userRes := utils.BuildUserResponse(&user)

	c.JSON(http.StatusCreated, utils.SuccessResponse{
		Status:  http.StatusCreated,
		Message: "User created successfully",
		Data: map[string]interface{}{
			"user": userRes,
		},
	})
}

func GetAllUsers(c *gin.Context, sqlDB *gorm.DB) {
	page := c.DefaultQuery("page", "1")
	limit := c.DefaultQuery("limit", "10")

	var users []models.User
	var total int64

	offset := (utils.StringToInt(page) - 1) * utils.StringToInt(limit)

	if err := sqlDB.Model(&models.User{}).Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: "Failed to count users",
		})
		return
	}
	if err := sqlDB.Offset(offset).Limit(utils.StringToInt(limit)).Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: "Failed to retrieve users",
		})
		return
	}

	roleIDSet := make(map[uint]bool)
	businessIDSet := make(map[uint]bool)
	for _, user := range users {
		roleIDSet[user.RoleID] = true
		businessIDSet[user.BusinessId] = true
	}

	roleIDs := make([]uint, 0, len(roleIDSet))
	for id := range roleIDSet {
		roleIDs = append(roleIDs, id)
	}

	businessIDs := make([]uint, 0, len(businessIDSet))
	for id := range businessIDSet {
		businessIDs = append(businessIDs, id)
	}
	var roles []models.Role
	var businesses []models.Business

	if len(roleIDs) > 0 {
		sqlDB.Select("id, name").Where("id IN ?", roleIDs).Find(&roles)
	}

	if len(businessIDs) > 0 {
		sqlDB.Select("id, name_en, name_ar").Where("id IN ?", businessIDs).Find(&businesses)
	}
	roleMap := make(map[uint]models.Role, len(roles))
	for _, role := range roles {
		roleMap[role.ID] = role
	}

	businessMap := make(map[uint]models.Business, len(businesses))
	for _, business := range businesses {
		businessMap[business.ID] = business
	}
	userList := make([]dtos.UserRes, 0, len(users))
	for _, user := range users {
		role := roleMap[user.RoleID]
		userList = append(userList, dtos.UserRes{
			ID:          user.ID,
			Email:       user.Email,
			FullName:    user.FullName,
			EmployeeId:  user.EmployeeId,
			Contact:     user.Contact,
			Gender:      user.Gender,
			DocumentId:  user.DocumentId,
			License:     user.License,
			Address:     user.Address,
			Nationality: user.Nationality,
			RoleID:      user.RoleID,
			RoleName:    role.Name,
			BranchId:    user.BranchId,
			BusinessId:  user.BusinessId,
			IsActive:    user.IsActive,
			IsStaff:     user.IsStaff,
			Service:     user.Service,
		})
	}

	c.JSON(http.StatusOK, utils.SuccessResponse{
		Status:  http.StatusOK,
		Message: "Users retrieved successfully",
		Data: map[string]interface{}{
			"users": userList,
			"pagination": map[string]interface{}{
				"total": total,
				"page":  utils.StringToInt(page),
				"limit": utils.StringToInt(limit),
			},
		},
	})
}
func GetUserByID(c *gin.Context, sqlDB *gorm.DB) {
	userID := c.Param("id")
	var user models.User
	if err := sqlDB.First(&user, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, utils.ErrorResponse{
				Status:  http.StatusNotFound,
				Message: "User not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: "Failed to retrieve user",
		})
		return
	}
	var role models.Role
	var business models.Business
	sqlDB.Select("id, name").First(&role, user.RoleID)
	sqlDB.Select("id, name_en, name_ar").First(&business, user.BusinessId)

	userRes := dtos.UserRes{
		ID:          user.ID,
		Email:       user.Email,
		FullName:    user.FullName,
		EmployeeId:  user.EmployeeId,
		Contact:     user.Contact,
		Gender:      user.Gender,
		DocumentId:  user.DocumentId,
		License:     user.License,
		Address:     user.Address,
		Nationality: user.Nationality,
		RoleID:      user.RoleID,
		RoleName:    role.Name,
		BranchId:    user.BranchId,
		BusinessId:  user.BusinessId,
		IsActive:    user.IsActive,
		IsStaff:     user.IsStaff,
		Service:     user.Service,
	}
	c.JSON(http.StatusOK, utils.SuccessResponse{
		Status:  http.StatusOK,
		Message: "User retrieved successfully",
		Data: map[string]interface{}{
			"user": userRes,
		},
	})
}
func UpdateUser(c *gin.Context, sqlDB *gorm.DB) {
	userID := c.Param("id")

	var req struct {
		FullName    string `json:"full_name"`
		Email       string `json:"email" binding:"omitempty,email"`
		EmployeeId  string `json:"employee_id"`
		Contact     string `json:"contact"`
		Gender      string `json:"gender"`
		RoleID      uint   `json:"role_id"`
		Address     string `json:"address"`
		Nationality string `json:"nationality"`
		DocumentId  string `json:"document_id"`
		License     string `json:"license"`
		BusinessId  uint   `json:"business_id"`
		BranchId    uint   `json:"branch_id"`
		IsActive    *bool  `json:"is_active"`
		IsStaff     *bool  `json:"is_staff"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Status:  http.StatusBadRequest,
			Message: "Invalid request: " + err.Error(),
		})
		return
	}
	var user models.User
	if err := sqlDB.First(&user, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, utils.ErrorResponse{
				Status:  http.StatusNotFound,
				Message: "User not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: "Failed to retrieve user",
		})
		return
	}
	if req.Email != "" && req.Email != user.Email {
		var existingUser models.User
		if err := sqlDB.Where("email = ? AND id != ?", req.Email, userID).First(&existingUser).Error; err == nil {
			c.JSON(http.StatusConflict, utils.ErrorResponse{
				Status:  http.StatusConflict,
				Message: "Email already exists",
			})
			return
		}
		user.Email = req.Email
	}
	if req.FullName != "" {
		user.FullName = req.FullName
	}
	if req.EmployeeId != "" {
		user.EmployeeId = req.EmployeeId
	}
	if req.Contact != "" {
		user.Contact = req.Contact
	}
	if req.Gender != "" {
		user.Gender = req.Gender
	}
	if req.RoleID != 0 {
		user.RoleID = req.RoleID
	}
	if req.Address != "" {
		user.Address = req.Address
	}
	if req.Nationality != "" {
		user.Nationality = req.Nationality
	}
	if req.DocumentId != "" {
		user.DocumentId = req.DocumentId
	}
	if req.License != "" {
		user.License = req.License
	}
	if req.BusinessId != 0 {
		user.BusinessId = req.BusinessId
	}
	if req.BranchId != 0 {
		user.BranchId = req.BranchId
	}
	if req.IsActive != nil {
		user.IsActive = *req.IsActive
	}
	if req.IsStaff != nil {
		user.IsStaff = *req.IsStaff
	}

	if err := sqlDB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: "Failed to update user: " + err.Error(),
		})
		return
	}

	var role models.Role
	var business models.Business
	sqlDB.Select("id, name").First(&role, user.RoleID)
	sqlDB.Select("id, name_en, name_ar").First(&business, user.BusinessId)

	userRes := dtos.UserRes{
		ID:          user.ID,
		Email:       user.Email,
		FullName:    user.FullName,
		EmployeeId:  user.EmployeeId,
		Contact:     user.Contact,
		Gender:      user.Gender,
		DocumentId:  user.DocumentId,
		License:     user.License,
		Address:     user.Address,
		Nationality: user.Nationality,
		RoleID:      user.RoleID,
		RoleName:    role.Name,
		BranchId:    user.BranchId,
		BusinessId:  user.BusinessId,
		IsActive:    user.IsActive,
		IsStaff:     user.IsStaff,
		Service:     user.Service,
	}

	c.JSON(http.StatusOK, utils.SuccessResponse{
		Status:  http.StatusOK,
		Message: "User updated successfully",
		Data: map[string]interface{}{
			"user": userRes,
		},
	})
}
func DeleteUser(c *gin.Context, sqlDB *gorm.DB) {
	userID := c.Param("id")

	var user models.User
	if err := sqlDB.First(&user, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, utils.ErrorResponse{
				Status:  http.StatusNotFound,
				Message: "User not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: "Failed to retrieve user",
		})
		return
	}

	if err := sqlDB.Delete(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: "Failed to delete user: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, utils.SuccessResponse{
		Status:  http.StatusOK,
		Message: "User deleted successfully",
		Data:    nil,
	})
}
func ChangePassword(c *gin.Context, sqlDB *gorm.DB) {

	var req dtos.ChangePasswordReq

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Status:  http.StatusBadRequest,
			Message: "Invalid request: " + err.Error(),
		})
		return
	}
	if req.OldPassword == req.NewPassword {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Status:  http.StatusBadRequest,
			Message: "New password cannot be the same as old password",
		})
		return
	}
	var user models.User
	if err := sqlDB.First(&user, req.UserID).Error; err != nil {
		c.JSON(http.StatusNotFound, utils.ErrorResponse{
			Status:  http.StatusNotFound,
			Message: "User not found",
		})
		return
	}

	if !utils.CheckPassword(req.OldPassword, user.Password) {
		c.JSON(http.StatusUnauthorized, utils.ErrorResponse{
			Status:  http.StatusUnauthorized,
			Message: "Old password is incorrect",
		})
		return
	}
	hashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: "Failed to hash password",
		})
		return
	}
	user.Password = hashedPassword
	if err := sqlDB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: "Failed to update password",
		})
		return
	}

	c.JSON(http.StatusOK, utils.SuccessResponse{
		Status:  http.StatusOK,
		Message: "Password changed successfully",
		Data: map[string]interface{}{
			"user_id": user.ID,
			"email":   user.Email,
		},
	})
}
func GetMyPermissions(c *gin.Context, sqlDB *gorm.DB) {
	userIDValue, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, utils.ErrorResponse{
			Status:  http.StatusUnauthorized,
			Message: "User not found in token",
		})
		return
	}

	userID, ok := userIDValue.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: "Invalid user ID format",
		})
		return
	}

	var user models.User
	if err := sqlDB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, utils.ErrorResponse{
			Status:  http.StatusNotFound,
			Message: "User not found",
		})
		return
	}

	var role models.Role
	if err := sqlDB.Select("id, name").First(&role, user.RoleID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: "Failed to retrieve role",
		})
		return
	}

	var permissions []models.Permission
	if err := sqlDB.Where("role_id = ?", user.RoleID).Find(&permissions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: "Failed to retrieve permissions",
		})
		return
	}

	moduleIDSet := make(map[uint]bool)
	for _, perm := range permissions {
		moduleIDSet[perm.ModuleID] = true
	}

	moduleIDs := make([]uint, 0, len(moduleIDSet))
	for id := range moduleIDSet {
		moduleIDs = append(moduleIDs, id)
	}
	var modules []models.Module
	if len(moduleIDs) > 0 {
		if err := sqlDB.Where("id IN ?", moduleIDs).Find(&modules).Error; err != nil {
			c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
				Status:  http.StatusInternalServerError,
				Message: "Failed to retrieve modules",
			})
			return
		}
	}

	moduleMap := make(map[uint]models.Module, len(modules))
	for _, module := range modules {
		moduleMap[module.ID] = module
	}

	modulePermissions := make(map[string][]string)
	permissionsList := make([]map[string]interface{}, 0, len(permissions))

	for _, perm := range permissions {
		module := moduleMap[perm.ModuleID]
		moduleName := module.Name

		if _, exists := modulePermissions[moduleName]; !exists {
			modulePermissions[moduleName] = []string{}
		}
		modulePermissions[moduleName] = append(modulePermissions[moduleName], perm.Action)

		permissionsList = append(permissionsList, map[string]interface{}{
			"id":          perm.ID,
			"module_id":   perm.ModuleID,
			"module_name": moduleName,
			"action":      perm.Action,
		})
	}

	c.JSON(http.StatusOK, utils.SuccessResponse{
		Status:  http.StatusOK,
		Message: "Permissions retrieved successfully",
		Data: map[string]interface{}{
			"user_id":            user.ID,
			"role_id":            user.RoleID,
			"role_name":          role.Name,
			"permissions":        permissionsList,
			"module_permissions": modulePermissions,
			"total_permissions":  len(permissionsList),
		},
	})
}

func CheckPermission(c *gin.Context, sqlDB *gorm.DB) {
	userIDValue, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, utils.ErrorResponse{
			Status:  http.StatusUnauthorized,
			Message: "User not found in token",
		})
		return
	}

	userID, ok := userIDValue.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: "Invalid user ID format",
		})
		return
	}

	var req struct {
		ModuleName string `json:"module_name" binding:"required"`
		Action     string `json:"action" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Status:  http.StatusBadRequest,
			Message: "Invalid request: " + err.Error(),
		})
		return
	}
	var user models.User
	if err := sqlDB.Select("id, role_id").First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, utils.ErrorResponse{
			Status:  http.StatusNotFound,
			Message: "User not found",
		})
		return
	}
	var module models.Module
	if err := sqlDB.Select("id").Where("name = ?", req.ModuleName).First(&module).Error; err != nil {
		c.JSON(http.StatusOK, utils.SuccessResponse{
			Status:  http.StatusOK,
			Message: "Permission check completed",
			Data: map[string]interface{}{
				"module_name":    req.ModuleName,
				"action":         req.Action,
				"has_permission": false,
			},
		})
		return
	}
	var count int64
	sqlDB.Model(&models.Permission{}).
		Where("role_id = ? AND module_id = ? AND action = ?", user.RoleID, module.ID, req.Action).
		Count(&count)

	c.JSON(http.StatusOK, utils.SuccessResponse{
		Status:  http.StatusOK,
		Message: "Permission check completed",
		Data: map[string]interface{}{
			"module_name":    req.ModuleName,
			"action":         req.Action,
			"has_permission": count > 0,
		},
	})
}
