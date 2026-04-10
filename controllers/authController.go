package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"supergit/inpatient/dtos"
	"supergit/inpatient/models"
	"supergit/inpatient/utils"
)

func Login(c *gin.Context, sqlDB *gorm.DB) {
	var req dtos.UserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Status:  http.StatusBadRequest,
			Message: "Invalid request format: " + err.Error(),
		})
		return
	}
	var user models.User
	if err := sqlDB.Where("email = ? AND is_active = ?", req.Email, true).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusUnauthorized, utils.ErrorResponse{
				Status:  http.StatusUnauthorized,
				Message: "Invalid email or password",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: "Database error",
		})
		return
	}
	if !utils.CheckPassword(req.Password, user.Password) {
		c.JSON(http.StatusUnauthorized, utils.ErrorResponse{
			Status:  http.StatusUnauthorized,
			Message: "Invalid email or password",
		})
		return
	}
	var business models.Business
	if err := sqlDB.First(&business, user.BusinessId).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: "Failed to retrieve business information",
		})
		return
	}
	var role models.Role
	if err := sqlDB.First(&role, user.RoleID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: "Failed to retrieve role information",
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
	token, err := utils.GenerateJWT(user.ID, user.RoleID, user.Email, business.InPatientDB, user.Service)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: "Failed to generate token",
		})
		return
	}

	modulePermissions := make(map[string][]map[string]interface{})
	for _, perm := range permissions {
		module := moduleMap[perm.ModuleID]
		moduleName := module.Name

		if _, exists := modulePermissions[moduleName]; !exists {
			modulePermissions[moduleName] = []map[string]interface{}{}
		}
		modulePermissions[moduleName] = append(modulePermissions[moduleName], map[string]interface{}{
			"id":         perm.ID,
			"permission": perm.Action,
			"check":      true,
			"module_id":  perm.ModuleID,
		})
	}

	userRes := dtos.UserRes{
		ID:          user.ID,
		Email:       user.Email,
		FullName:    user.FullName,
		EmployeeId:  user.EmployeeId,
		Contact:     user.Contact,
		DocumentId:  user.DocumentId,
		Address:     user.Address,
		Nationality: user.Nationality,
		RoleID:      user.RoleID,
		RoleName:    role.Name,
		BranchId:    user.BranchId,
		BusinessId:  user.BusinessId,
		FCMToken:    user.FCMToken,
		IsActive:    user.IsActive,
		Service:     user.Service,
	}

	c.JSON(http.StatusOK, utils.SuccessResponse{
		Status:  http.StatusOK,
		Message: "Login successful",
		Data: map[string]interface{}{
			"user":        userRes,
			"token":       token,
			"role_name":   role.Name,
			"permissions": modulePermissions,
			"business":    business,
		},
	})
}

func GetUser(c *gin.Context, sqlDB *gorm.DB) {
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
	var business models.Business
	if err := sqlDB.First(&business, user.BusinessId).Error; err != nil {
		business = models.Business{}
	}

	userRes := dtos.UserRes{
		ID:          user.ID,
		Email:       user.Email,
		FullName:    user.FullName,
		EmployeeId:  user.EmployeeId,
		Contact:     user.Contact,
		DocumentId:  user.DocumentId,
		Address:     user.Address,
		Nationality: user.Nationality,
		RoleID:      user.RoleID,
		RoleName:    role.Name,
		BranchId:    user.BranchId,
		BusinessId:  user.BusinessId,
		FCMToken:    user.FCMToken,
		IsActive:    user.IsActive,
		Service:     user.Service,
	}

	c.JSON(http.StatusOK, utils.SuccessResponse{
		Status:  http.StatusOK,
		Message: "User retrieved successfully",
		Data: map[string]interface{}{
			"user":      userRes,
			"role_name": role.Name,
			"business":  business,
		},
	})
}

func Logout(c *gin.Context, sqlDB *gorm.DB) {
	userIDValue, exists := c.Get("user_id")
	if exists {
		userID, ok := userIDValue.(uint)
		if ok {
			_ = userID
		}
	}
	c.JSON(http.StatusOK, utils.SuccessResponse{
		Status:  http.StatusOK,
		Message: "Logout successful. Please clear your local storage and token.",
		Data: map[string]interface{}{
			"logout": true,
			"instructions": map[string]string{
				"action": "clear_storage",
				"items":  "token, user, permissions, business, subscription",
			},
		},
	})
}
