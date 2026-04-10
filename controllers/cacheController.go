package controllers

import (
	"net/http"

	"supergit/inpatient/middleware"
	"supergit/inpatient/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RefreshPermissionCacheHandler(c *gin.Context, sqlDB *gorm.DB) {
	if err := middleware.RefreshPermissionCache(sqlDB); err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: "Failed to refresh permission cache: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, utils.SuccessResponse{
		Status:  http.StatusOK,
		Message: "Permission cache refreshed successfully",
		Data:    nil,
	})
}
