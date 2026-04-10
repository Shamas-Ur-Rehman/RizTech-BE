package controllers

import (
	"net/http"

	"supergit/inpatient/dtos"
	"supergit/inpatient/models"
	"supergit/inpatient/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func CreatePermission(c *gin.Context, sqlDB *gorm.DB) {
	var req dtos.PermissionReq

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Status:  http.StatusBadRequest,
			Message: "Invalid request: " + err.Error(),
		})
		return
	}
	validActions := []string{"get", "create", "update", "delete"}
	isValidAction := false
	for _, action := range validActions {
		if req.Action == action {
			isValidAction = true
			break
		}
	}
	if !isValidAction {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Status:  http.StatusBadRequest,
			Message: "Invalid action. Must be one of: get, create, update, delete",
		})
		return
	}
	var role models.Role
	if err := sqlDB.Select("id, name").First(&role, req.RoleID).Error; err != nil {
		c.JSON(http.StatusNotFound, utils.ErrorResponse{
			Status:  http.StatusNotFound,
			Message: "Role not found",
		})
		return
	}
	var module models.Module
	if err := sqlDB.Select("id, name").First(&module, req.ModuleID).Error; err != nil {
		c.JSON(http.StatusNotFound, utils.ErrorResponse{
			Status:  http.StatusNotFound,
			Message: "Module not found",
		})
		return
	}
	var existingPerm models.Permission
	if err := sqlDB.Where("role_id = ? AND module_id = ? AND action = ?", req.RoleID, req.ModuleID, req.Action).First(&existingPerm).Error; err == nil {
		c.JSON(http.StatusConflict, utils.ErrorResponse{
			Status:  http.StatusConflict,
			Message: "Permission already exists for this role, module, and action",
		})
		return
	}

	permission := models.Permission{
		RoleID:     req.RoleID,
		ModuleID:   req.ModuleID,
		Action:     req.Action,
		BusinessId: req.BusinessId,
		BranchId:   req.BranchId,
	}

	if err := sqlDB.Create(&permission).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: "Failed to create permission: " + err.Error(),
		})
		return
	}

	permRes := dtos.PermissionRes{
		ID:         permission.ID,
		RoleID:     permission.RoleID,
		RoleName:   role.Name,
		ModuleID:   permission.ModuleID,
		ModuleName: module.Name,
		Action:     permission.Action,
		BusinessId: permission.BusinessId,
		BranchId:   permission.BranchId,
		CreatedAt:  permission.CreatedAt,
		UpdatedAt:  permission.UpdatedAt,
	}

	c.JSON(http.StatusCreated, utils.SuccessResponse{
		Status:  http.StatusCreated,
		Message: "Permission created successfully",
		Data: map[string]interface{}{
			"permission": permRes,
		},
	})
}

func GetAllPermissions(c *gin.Context, sqlDB *gorm.DB) {
	var permissions []models.Permission

	if err := sqlDB.Find(&permissions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: "Failed to retrieve permissions",
		})
		return
	}
	roleIDSet := make(map[uint]bool)
	moduleIDSet := make(map[uint]bool)
	for _, perm := range permissions {
		roleIDSet[perm.RoleID] = true
		moduleIDSet[perm.ModuleID] = true
	}

	roleIDs := make([]uint, 0, len(roleIDSet))
	for id := range roleIDSet {
		roleIDs = append(roleIDs, id)
	}

	moduleIDs := make([]uint, 0, len(moduleIDSet))
	for id := range moduleIDSet {
		moduleIDs = append(moduleIDs, id)
	}
	var roles []models.Role
	var modules []models.Module

	if len(roleIDs) > 0 {
		sqlDB.Select("id, name").Where("id IN ?", roleIDs).Find(&roles)
	}

	if len(moduleIDs) > 0 {
		sqlDB.Select("id, name").Where("id IN ?", moduleIDs).Find(&modules)
	}
	roleMap := make(map[uint]models.Role, len(roles))
	for _, role := range roles {
		roleMap[role.ID] = role
	}

	moduleMap := make(map[uint]models.Module, len(modules))
	for _, module := range modules {
		moduleMap[module.ID] = module
	}
	permList := make([]dtos.PermissionRes, 0, len(permissions))
	for _, perm := range permissions {
		role := roleMap[perm.RoleID]
		module := moduleMap[perm.ModuleID]
		permList = append(permList, dtos.PermissionRes{
			ID:         perm.ID,
			RoleID:     perm.RoleID,
			RoleName:   role.Name,
			ModuleID:   perm.ModuleID,
			ModuleName: module.Name,
			Action:     perm.Action,
			BusinessId: perm.BusinessId,
			BranchId:   perm.BranchId,
			CreatedAt:  perm.CreatedAt,
			UpdatedAt:  perm.UpdatedAt,
		})
	}

	c.JSON(http.StatusOK, utils.SuccessResponse{
		Status:  http.StatusOK,
		Message: "Permissions retrieved successfully",
		Data: map[string]interface{}{
			"permissions": permList,
			"total":       len(permList),
		},
	})
}

func GetPermissionByID(c *gin.Context, sqlDB *gorm.DB) {
	permissionID := c.Param("id")

	var permission models.Permission
	if err := sqlDB.First(&permission, permissionID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, utils.ErrorResponse{
				Status:  http.StatusNotFound,
				Message: "Permission not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: "Failed to retrieve permission",
		})
		return
	}
	var role models.Role
	var module models.Module
	sqlDB.Select("id, name").First(&role, permission.RoleID)
	sqlDB.Select("id, name").First(&module, permission.ModuleID)

	permRes := dtos.PermissionRes{
		ID:         permission.ID,
		RoleID:     permission.RoleID,
		RoleName:   role.Name,
		ModuleID:   permission.ModuleID,
		ModuleName: module.Name,
		Action:     permission.Action,
		BusinessId: permission.BusinessId,
		BranchId:   permission.BranchId,
		CreatedAt:  permission.CreatedAt,
		UpdatedAt:  permission.UpdatedAt,
	}

	c.JSON(http.StatusOK, utils.SuccessResponse{
		Status:  http.StatusOK,
		Message: "Permission retrieved successfully",
		Data: map[string]interface{}{
			"permission": permRes,
		},
	})
}

func UpdatePermission(c *gin.Context, sqlDB *gorm.DB) {
	permissionID := c.Param("id")

	var req struct {
		Action     string `json:"action"`
		BusinessId uint   `json:"business_id"`
		BranchId   uint   `json:"branch_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Status:  http.StatusBadRequest,
			Message: "Invalid request: " + err.Error(),
		})
		return
	}

	var permission models.Permission
	if err := sqlDB.First(&permission, permissionID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, utils.ErrorResponse{
				Status:  http.StatusNotFound,
				Message: "Permission not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: "Failed to retrieve permission",
		})
		return
	}

	if req.Action != "" {
		validActions := []string{"get", "create", "update", "delete"}
		isValidAction := false
		for _, action := range validActions {
			if req.Action == action {
				isValidAction = true
				break
			}
		}
		if !isValidAction {
			c.JSON(http.StatusBadRequest, utils.ErrorResponse{
				Status:  http.StatusBadRequest,
				Message: "Invalid action. Must be one of: get, create, update, delete",
			})
			return
		}
		var existingPerm models.Permission
		if err := sqlDB.Where("role_id = ? AND module_id = ? AND action = ? AND id != ?", permission.RoleID, permission.ModuleID, req.Action, permissionID).First(&existingPerm).Error; err == nil {
			c.JSON(http.StatusConflict, utils.ErrorResponse{
				Status:  http.StatusConflict,
				Message: "Permission with this action already exists for this role and module",
			})
			return
		}

		permission.Action = req.Action
	}
	if req.BusinessId != 0 {
		permission.BusinessId = req.BusinessId
	}
	if req.BranchId != 0 {
		permission.BranchId = req.BranchId
	}

	if err := sqlDB.Save(&permission).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: "Failed to update permission: " + err.Error(),
		})
		return
	}
	var role models.Role
	var module models.Module
	sqlDB.Select("id, name").First(&role, permission.RoleID)
	sqlDB.Select("id, name").First(&module, permission.ModuleID)

	permRes := dtos.PermissionRes{
		ID:         permission.ID,
		RoleID:     permission.RoleID,
		RoleName:   role.Name,
		ModuleID:   permission.ModuleID,
		ModuleName: module.Name,
		Action:     permission.Action,
		BusinessId: permission.BusinessId,
		BranchId:   permission.BranchId,
		CreatedAt:  permission.CreatedAt,
		UpdatedAt:  permission.UpdatedAt,
	}

	c.JSON(http.StatusOK, utils.SuccessResponse{
		Status:  http.StatusOK,
		Message: "Permission updated successfully",
		Data: map[string]interface{}{
			"permission": permRes,
		},
	})
}

func DeletePermission(c *gin.Context, sqlDB *gorm.DB) {
	permissionID := c.Param("id")

	var permission models.Permission
	if err := sqlDB.First(&permission, permissionID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, utils.ErrorResponse{
				Status:  http.StatusNotFound,
				Message: "Permission not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: "Failed to retrieve permission",
		})
		return
	}
	if err := sqlDB.Delete(&permission).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: "Failed to delete permission: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, utils.SuccessResponse{
		Status:  http.StatusOK,
		Message: "Permission deleted successfully",
		Data:    nil,
	})
}
