package controllers

import (
	"fmt"
	"net/http"

	"supergit/inpatient/dtos"
	"supergit/inpatient/middleware"
	"supergit/inpatient/models"
	"supergit/inpatient/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func AssignPermissionsToRole(c *gin.Context, sqlDB *gorm.DB) {
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

	var req dtos.AssignPermissionsReq

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Status:  http.StatusBadRequest,
			Message: "Invalid request: " + err.Error(),
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
	if err := sqlDB.First(&role, req.RoleID).Error; err != nil {
		c.JSON(http.StatusNotFound, utils.ErrorResponse{
			Status:  http.StatusNotFound,
			Message: "Role not found",
		})
		return
	}

	validActions := map[string]bool{"get": true, "create": true, "update": true, "delete": true}

	moduleIDs := make([]uint, 0, len(req.Permissions))
	for _, modulePerm := range req.Permissions {
		moduleIDs = append(moduleIDs, modulePerm.ModuleID)
	}
	var modules []models.Module
	if err := sqlDB.Where("id IN ?", moduleIDs).Find(&modules).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: "Failed to fetch modules: " + err.Error(),
		})
		return
	}
	moduleMap := make(map[uint]models.Module, len(modules))
	for _, module := range modules {
		moduleMap[module.ID] = module
	}
	for _, modulePerm := range req.Permissions {
		if _, exists := moduleMap[modulePerm.ModuleID]; !exists {
			c.JSON(http.StatusNotFound, utils.ErrorResponse{
				Status:  http.StatusNotFound,
				Message: fmt.Sprintf("Module not found with ID: %d", modulePerm.ModuleID),
			})
			return
		}
	}
	var existingPermissions []models.Permission
	sqlDB.Where("role_id = ?", req.RoleID).Find(&existingPermissions)
	existingPermMap := make(map[string]models.Permission)
	for _, perm := range existingPermissions {
		key := fmt.Sprintf("%d-%s", perm.ModuleID, perm.Action)
		existingPermMap[key] = perm
	}

	shouldExistMap := make(map[string]bool)
	permissionsToCreate := []models.Permission{}
	upsertedPermissions := []dtos.PermissionRes{}

	for _, modulePerm := range req.Permissions {
		module := moduleMap[modulePerm.ModuleID]
		for _, action := range modulePerm.Actions {
			if !validActions[action] {
				c.JSON(http.StatusBadRequest, utils.ErrorResponse{
					Status:  http.StatusBadRequest,
					Message: "Invalid action: " + action + ". Must be one of: get, create, update, delete",
				})
				return
			}

			key := fmt.Sprintf("%d-%s", modulePerm.ModuleID, action)
			shouldExistMap[key] = true

			if existingPerm, exists := existingPermMap[key]; exists {
				upsertedPermissions = append(upsertedPermissions, dtos.PermissionRes{
					ID:         existingPerm.ID,
					RoleID:     existingPerm.RoleID,
					ModuleID:   existingPerm.ModuleID,
					ModuleName: module.Name,
					Action:     existingPerm.Action,
					BusinessId: existingPerm.BusinessId,
					BranchId:   existingPerm.BranchId,
					CreatedAt:  existingPerm.CreatedAt,
					UpdatedAt:  existingPerm.UpdatedAt,
				})
			} else {
				permissionsToCreate = append(permissionsToCreate, models.Permission{
					RoleID:     req.RoleID,
					ModuleID:   modulePerm.ModuleID,
					Action:     action,
					BusinessId: user.BusinessId,
					BranchId:   user.BranchId,
				})
			}
		}
	}
	if len(permissionsToCreate) > 0 {
		if err := sqlDB.CreateInBatches(permissionsToCreate, 100).Error; err != nil {
			c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
				Status:  http.StatusInternalServerError,
				Message: "Failed to create permissions: " + err.Error(),
			})
			return
		}
		for _, perm := range permissionsToCreate {
			module := moduleMap[perm.ModuleID]
			upsertedPermissions = append(upsertedPermissions, dtos.PermissionRes{
				ID:         perm.ID,
				RoleID:     perm.RoleID,
				ModuleID:   perm.ModuleID,
				ModuleName: module.Name,
				Action:     perm.Action,
				BusinessId: perm.BusinessId,
				BranchId:   perm.BranchId,
				CreatedAt:  perm.CreatedAt,
				UpdatedAt:  perm.UpdatedAt,
			})
		}
	}
	permissionsToDelete := []uint{}
	for key, perm := range existingPermMap {
		if !shouldExistMap[key] {
			permissionsToDelete = append(permissionsToDelete, perm.ID)
		}
	}

	if len(permissionsToDelete) > 0 {
		sqlDB.Where("id IN ?", permissionsToDelete).Delete(&models.Permission{})
	}

	go middleware.RefreshPermissionCache(sqlDB)

	c.JSON(http.StatusOK, utils.SuccessResponse{
		Status:  http.StatusOK,
		Message: "Permissions assigned to role successfully",
		Data: map[string]interface{}{
			"role_id":     role.ID,
			"role_name":   role.Name,
			"permissions": upsertedPermissions,
			"total":       len(upsertedPermissions),
			"deleted":     len(permissionsToDelete),
		},
	})
}

func RemovePermissionsFromRole(c *gin.Context, sqlDB *gorm.DB) {
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

	var req dtos.RemovePermissionsReq

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Status:  http.StatusBadRequest,
			Message: "Invalid request: " + err.Error(),
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
	if err := sqlDB.First(&role, req.RoleID).Error; err != nil {
		c.JSON(http.StatusNotFound, utils.ErrorResponse{
			Status:  http.StatusNotFound,
			Message: "Role not found",
		})
		return
	}

	if err := sqlDB.Where("id IN ? AND role_id = ?", req.PermissionIDs, req.RoleID).Delete(&models.Permission{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: "Failed to remove permissions: " + err.Error(),
		})
		return
	}

	go middleware.RefreshPermissionCache(sqlDB)
	var permissions []models.Permission
	sqlDB.Where("role_id = ?", req.RoleID).Find(&permissions)
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
		sqlDB.Where("id IN ?", moduleIDs).Find(&modules)
	}
	moduleMap := make(map[uint]models.Module, len(modules))
	for _, module := range modules {
		moduleMap[module.ID] = module
	}

	permList := make([]dtos.PermissionRes, 0, len(permissions))
	for _, perm := range permissions {
		module := moduleMap[perm.ModuleID]
		permList = append(permList, dtos.PermissionRes{
			ID:         perm.ID,
			RoleID:     perm.RoleID,
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
		Message: "Permissions removed from role successfully",
		Data: map[string]interface{}{
			"role_id":     role.ID,
			"role_name":   role.Name,
			"permissions": permList,
		},
	})
}

func GetPermissionsByRoleID(c *gin.Context, sqlDB *gorm.DB) {
	roleID := c.Param("id")
	var role models.Role
	if err := sqlDB.First(&role, roleID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, utils.ErrorResponse{
				Status:  http.StatusNotFound,
				Message: "Role not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: "Failed to retrieve role",
		})
		return
	}
	var permissions []models.Permission
	if err := sqlDB.Where("role_id = ?", roleID).Find(&permissions).Error; err != nil {
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
	if err := sqlDB.Where("id IN ?", moduleIDs).Find(&modules).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: "Failed to retrieve modules",
		})
		return
	}
	moduleMap := make(map[uint]models.Module, len(modules))
	for _, module := range modules {
		moduleMap[module.ID] = module
	}
	modulePermissionsMap := make(map[uint]map[string]interface{})
	permissionsList := make([]dtos.PermissionRes, 0, len(permissions))

	for _, perm := range permissions {
		module := moduleMap[perm.ModuleID]
		permissionsList = append(permissionsList, dtos.PermissionRes{
			ID:         perm.ID,
			RoleID:     perm.RoleID,
			ModuleID:   perm.ModuleID,
			ModuleName: module.Name,
			Action:     perm.Action,
			BusinessId: perm.BusinessId,
			BranchId:   perm.BranchId,
			CreatedAt:  perm.CreatedAt,
			UpdatedAt:  perm.UpdatedAt,
		})

		if _, exists := modulePermissionsMap[perm.ModuleID]; !exists {
			modulePermissionsMap[perm.ModuleID] = map[string]interface{}{
				"module_id":   perm.ModuleID,
				"module_name": module.Name,
				"actions":     []string{},
			}
		}

		actions := modulePermissionsMap[perm.ModuleID]["actions"].([]string)
		actions = append(actions, perm.Action)
		modulePermissionsMap[perm.ModuleID]["actions"] = actions
	}

	groupedPermissions := make([]map[string]interface{}, 0, len(modulePermissionsMap))
	for _, moduleData := range modulePermissionsMap {
		groupedPermissions = append(groupedPermissions, moduleData)
	}

	c.JSON(http.StatusOK, utils.SuccessResponse{
		Status:  http.StatusOK,
		Message: "Permissions retrieved successfully",
		Data: map[string]interface{}{
			"role_id":             role.ID,
			"role_name":           role.Name,
			"permissions":         permissionsList,
			"grouped_permissions": groupedPermissions,
			"total":               len(permissionsList),
		},
	})
}
