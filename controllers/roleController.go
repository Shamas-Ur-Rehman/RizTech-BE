package controllers

import (
	"net/http"
	"supergit/inpatient/dtos"
	"supergit/inpatient/models"
	"supergit/inpatient/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func CreateRole(c *gin.Context, sqlDB *gorm.DB) {
	var req dtos.RoleReq

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Status:  http.StatusBadRequest,
			Message: "Invalid request: " + err.Error(),
		})
		return
	}
	var existingRole models.Role
	if err := sqlDB.Where("LOWER(name) = LOWER(?)", req.Name).First(&existingRole).Error; err == nil {
		c.JSON(http.StatusConflict, utils.ErrorResponse{
			Status:  http.StatusConflict,
			Message: "Role name already exists (case-insensitive)",
		})
		return
	}

	role := models.Role{
		Name:       req.Name,
		BusinessId: req.BusinessId,
		BranchId:   req.BranchId,
	}

	if err := sqlDB.Create(&role).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: "Failed to create role: " + err.Error(),
		})
		return
	}

	roleRes := dtos.RoleRes{
		ID:         role.ID,
		Name:       role.Name,
		BusinessId: role.BusinessId,
		BranchId:   role.BranchId,
		CreatedAt:  role.CreatedAt,
		UpdatedAt:  role.UpdatedAt,
	}

	c.JSON(http.StatusCreated, utils.SuccessResponse{
		Status:  http.StatusCreated,
		Message: "Role created successfully",
		Data: map[string]interface{}{
			"role": roleRes,
		},
	})
}

func GetAllRoles(c *gin.Context, sqlDB *gorm.DB) {
	page := utils.StringToInt(c.DefaultQuery("page", "1"))
	perPage := utils.StringToInt(c.DefaultQuery("per_page", "10"))
	search := c.Query("search")

	query := sqlDB.Model(&models.Role{})

	if search != "" {
		query = query.Where("LOWER(name) LIKE LOWER(?)", "%"+search+"%")
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: "Failed to count roles",
		})
		return
	}

	offset := (page - 1) * perPage
	totalPages := int((total + int64(perPage) - 1) / int64(perPage))

	var roles []models.Role
	if err := query.Offset(offset).Limit(perPage).Order("created_at DESC").Find(&roles).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: "Failed to retrieve roles",
		})
		return
	}

	roleIDs := make([]uint, len(roles))
	for i, role := range roles {
		roleIDs[i] = role.ID
	}

	var permissions []models.Permission
	if len(roleIDs) > 0 {
		if err := sqlDB.Where("role_id IN ?", roleIDs).Find(&permissions).Error; err != nil {
			c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
				Status:  http.StatusInternalServerError,
				Message: "Failed to retrieve permissions",
			})
			return
		}
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

	rolePermissionsMap := make(map[uint][]models.Permission)
	for _, perm := range permissions {
		rolePermissionsMap[perm.RoleID] = append(rolePermissionsMap[perm.RoleID], perm)
	}

	roleList := make([]dtos.RoleRes, 0, len(roles))
	for _, role := range roles {
		permList := make([]dtos.PermissionRes, 0)

		if perms, exists := rolePermissionsMap[role.ID]; exists {
			for _, perm := range perms {
				module := moduleMap[perm.ModuleID]
				permList = append(permList, dtos.PermissionRes{
					ID:         perm.ID,
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

		roleList = append(roleList, dtos.RoleRes{
			ID:          role.ID,
			Name:        role.Name,
			BusinessId:  role.BusinessId,
			BranchId:    role.BranchId,
			CreatedAt:   role.CreatedAt,
			UpdatedAt:   role.UpdatedAt,
			Permissions: permList,
		})
	}

	c.JSON(http.StatusOK, utils.SuccessResponse{
		Status:  http.StatusOK,
		Message: "Roles retrieved successfully",
		Data: map[string]interface{}{
			"roles": roleList,
			"pagination": map[string]interface{}{
				"total_records": total,
				"total_pages":   totalPages,
				"page":          page,
				"per_page":      perPage,
			},
		},
	})
}
func GetAllRolesList(c *gin.Context, sqlDB *gorm.DB) {
	search := c.Query("search")

	query := sqlDB.Model(&models.Role{}).Where("deleted_at IS NULL")

	if search != "" {
		query = query.Where("LOWER(name) LIKE LOWER(?)", "%"+search+"%")
	}

	var roles []models.Role
	if err := query.Order("created_at DESC").Find(&roles).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: "Failed to retrieve roles",
		})
		return
	}

	c.JSON(http.StatusOK, utils.SuccessResponse{
		Status:  http.StatusOK,
		Message: "Roles retrieved successfully",
		Data:    roles,
	})
}

func GetRoleByID(c *gin.Context, sqlDB *gorm.DB) {
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
	permList := make([]dtos.PermissionRes, 0, len(permissions))
	for _, perm := range permissions {
		module := moduleMap[perm.ModuleID]
		permList = append(permList, dtos.PermissionRes{
			ID:         perm.ID,
			ModuleID:   perm.ModuleID,
			ModuleName: module.Name,
			Action:     perm.Action,
			BusinessId: perm.BusinessId,
			BranchId:   perm.BranchId,
			CreatedAt:  perm.CreatedAt,
			UpdatedAt:  perm.UpdatedAt,
		})
	}

	roleRes := dtos.RoleRes{
		ID:          role.ID,
		Name:        role.Name,
		BusinessId:  role.BusinessId,
		BranchId:    role.BranchId,
		CreatedAt:   role.CreatedAt,
		UpdatedAt:   role.UpdatedAt,
		Permissions: permList,
	}

	c.JSON(http.StatusOK, utils.SuccessResponse{
		Status:  http.StatusOK,
		Message: "Role retrieved successfully",
		Data: map[string]interface{}{
			"role": roleRes,
		},
	})
}
func UpdateRole(c *gin.Context, sqlDB *gorm.DB) {
	roleID := c.Param("id")

	var req dtos.RoleUpdateReq

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Status:  http.StatusBadRequest,
			Message: "Invalid request: " + err.Error(),
		})
		return
	}

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

	if req.Name != "" {
		var existingRole models.Role
		if err := sqlDB.Where("LOWER(name) = LOWER(?) AND id != ?", req.Name, roleID).First(&existingRole).Error; err == nil {
			c.JSON(http.StatusConflict, utils.ErrorResponse{
				Status:  http.StatusConflict,
				Message: "Role name already exists (case-insensitive)",
			})
			return
		}
		role.Name = req.Name
	}
	if req.BusinessId != 0 {
		role.BusinessId = req.BusinessId
	}
	if req.BranchId != 0 {
		role.BranchId = req.BranchId
	}

	if err := sqlDB.Save(&role).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: "Failed to update role: " + err.Error(),
		})
		return
	}

	roleRes := dtos.RoleRes{
		ID:         role.ID,
		Name:       role.Name,
		BusinessId: role.BusinessId,
		BranchId:   role.BranchId,
		CreatedAt:  role.CreatedAt,
		UpdatedAt:  role.UpdatedAt,
	}

	c.JSON(http.StatusOK, utils.SuccessResponse{
		Status:  http.StatusOK,
		Message: "Role updated successfully",
		Data: map[string]interface{}{
			"role": roleRes,
		},
	})
}
func DeleteRole(c *gin.Context, sqlDB *gorm.DB) {
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
	var userCount int64
	if err := sqlDB.Model(&models.User{}).Where("role_id = ?", roleID).Count(&userCount).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: "Failed to check role usage",
		})
		return
	}

	if userCount > 0 {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Status:  http.StatusBadRequest,
			Message: "Cannot delete role that is assigned to users",
		})
		return
	}
	if err := sqlDB.Delete(&role).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: "Failed to delete role: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, utils.SuccessResponse{
		Status:  http.StatusOK,
		Message: "Role deleted successfully",
		Data:    nil,
	})
}
