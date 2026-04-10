package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"supergit/inpatient/models"
	"supergit/inpatient/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type permissionCache struct {
	permissions map[string]bool
	userActive  map[uint]bool
	userRole    map[uint]uint
	mu          sync.RWMutex
	lastRefresh time.Time
	ttl         time.Duration
}

var cache = &permissionCache{
	permissions: make(map[string]bool),
	userActive:  make(map[uint]bool),
	userRole:    make(map[uint]uint),
	ttl:         10 * time.Minute,
}

func RefreshPermissionCache(db *gorm.DB) error {
	cache.mu.Lock()
	defer cache.mu.Unlock()
	var users []models.User
	if err := db.Select("id, role_id, is_active").Find(&users).Error; err != nil {
		return err
	}

	newUserActive := make(map[uint]bool, len(users))
	newUserRole := make(map[uint]uint, len(users))
	for _, user := range users {
		newUserActive[user.ID] = user.IsActive
		newUserRole[user.ID] = user.RoleID
	}
	var permissions []models.Permission
	if err := db.Select("role_id, module_id, action").Find(&permissions).Error; err != nil {
		return err
	}
	moduleIDs := make([]uint, 0, len(permissions))
	moduleIDSet := make(map[uint]bool)
	for _, perm := range permissions {
		if !moduleIDSet[perm.ModuleID] {
			moduleIDs = append(moduleIDs, perm.ModuleID)
			moduleIDSet[perm.ModuleID] = true
		}
	}
	var modules []models.Module
	if err := db.Select("id, name").Where("id IN ?", moduleIDs).Find(&modules).Error; err != nil {
		return err
	}
	moduleMap := make(map[uint]string, len(modules))
	for _, module := range modules {
		moduleMap[module.ID] = module.Name
	}
	newPermissions := make(map[string]bool, len(permissions))
	for _, perm := range permissions {
		moduleName := moduleMap[perm.ModuleID]
		key := fmt.Sprintf("%d-%s-%s", perm.RoleID, moduleName, perm.Action)
		newPermissions[key] = true
	}

	cache.permissions = newPermissions
	cache.userActive = newUserActive
	cache.userRole = newUserRole
	cache.lastRefresh = time.Now()

	return nil
}
func (pc *permissionCache) needsRefresh() bool {
	pc.mu.RLock()
	defer pc.mu.RUnlock()
	return time.Since(pc.lastRefresh) > pc.ttl
}
func (pc *permissionCache) hasPermission(userID uint, moduleName, action string) (bool, bool, error) {
	pc.mu.RLock()
	defer pc.mu.RUnlock()
	isActive, userExists := pc.userActive[userID]
	if !userExists {
		return false, false, fmt.Errorf("user not found in cache")
	}
	if !isActive {
		return false, false, fmt.Errorf("user is inactive")
	}
	roleID, roleExists := pc.userRole[userID]
	if !roleExists {
		return false, false, fmt.Errorf("user role not found in cache")
	}
	key := fmt.Sprintf("%d-%s-%s", roleID, moduleName, action)
	hasPermission := pc.permissions[key]

	return hasPermission, true, nil
}

func RolePermissionMiddleware(moduleName, action string, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDValue, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, utils.ErrorResponse{
				Status:  http.StatusUnauthorized,
				Message: "User not authenticated",
			})
			c.Abort()
			return
		}

		userID, ok := userIDValue.(uint)
		if !ok {
			c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
				Status:  http.StatusInternalServerError,
				Message: "Invalid user ID format",
			})
			c.Abort()
			return
		}
		if cache.needsRefresh() {
			go RefreshPermissionCache(db)
		}
		hasPermission, found, err := cache.hasPermission(userID, moduleName, action)
		if !found || err != nil {
			var user models.User
			if err := db.Select("id, role_id, is_active").First(&user, userID).Error; err != nil {
				c.JSON(http.StatusNotFound, utils.ErrorResponse{
					Status:  http.StatusNotFound,
					Message: "User not found",
				})
				c.Abort()
				return
			}

			if !user.IsActive {
				c.JSON(http.StatusForbidden, utils.ErrorResponse{
					Status:  http.StatusForbidden,
					Message: "User account is inactive",
				})
				c.Abort()
				return
			}
			var module models.Module
			if err := db.Select("id").Where("name = ?", moduleName).First(&module).Error; err != nil {
				c.JSON(http.StatusForbidden, utils.ErrorResponse{
					Status:  http.StatusForbidden,
					Message: "Module not found",
				})
				c.Abort()
				return
			}

			var count int64
			db.Model(&models.Permission{}).
				Where("role_id = ? AND module_id = ? AND action = ?", user.RoleID, module.ID, action).
				Count(&count)

			hasPermission = count > 0
		}

		if err != nil && err.Error() == "user is inactive" {
			c.JSON(http.StatusForbidden, utils.ErrorResponse{
				Status:  http.StatusForbidden,
				Message: "User account is inactive",
			})
			c.Abort()
			return
		}

		if !hasPermission {
			c.JSON(http.StatusForbidden, utils.ErrorResponse{
				Status:  http.StatusForbidden,
				Message: "Insufficient permissions for this action",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
