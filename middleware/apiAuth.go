package middleware

import (
	"fmt"
	"net/http"

	"supergit/inpatient/config"
	"supergit/inpatient/models"
	"supergit/inpatient/utils"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"
)
func APIKeyMiddleware(sqlDB *gorm.DB, mongoClient *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			c.JSON(http.StatusUnauthorized, utils.ErrorResponse{
				Status:  http.StatusUnauthorized,
				Message: "API key is required",
			})
			c.Abort()
			return
		}
		var subscription models.Subscription
		if err := sqlDB.Where("api_key = ?", apiKey).First(&subscription).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusUnauthorized, utils.ErrorResponse{
					Status:  http.StatusUnauthorized,
					Message: "Invalid API key",
				})
				c.Abort()
				return
			}
			c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
				Status:  http.StatusInternalServerError,
				Message: "Failed to validate API key",
			})
			c.Abort()
			return
		}

		businessID := subscription.BusinessId
		branchID := subscription.BranchId
		var business models.Business
		if err := sqlDB.Select("id", "in_patient_db", "name_en").First(&business, businessID).Error; err != nil {
			c.JSON(http.StatusNotFound, utils.ErrorResponse{
				Status:  http.StatusNotFound,
				Message: "Business not found for this subscription",
			})
			c.Abort()
			return
		}
		dbName := business.InPatientDB
		if dbName == "" {
			dbName = "inpatient_main_db"
		}
		employeeID := fmt.Sprintf("API-BUS-%d", businessID)
		employeeName := fmt.Sprintf("API: %s", business.NameEn)

		virtualUser := &models.User{
			ID:         0, 
			FullName:   employeeName,
			EmployeeId: employeeID,
			BusinessId: businessID,
			BranchId:   branchID,
			RoleID:     0,
			IsActive:   true,
		}
		c.Set("business_id", businessID)
		c.Set("branch_id", branchID)
		c.Set("tenant_db", dbName)
		c.Set("business", business)
		c.Set("user", virtualUser)
		c.Set("subscription", subscription)
		c.Set("is_api_request", true)
		c.Set("user_id", fmt.Sprintf("api-%d", businessID))

		collections := config.GetCollections(mongoClient, dbName)
		c.Set("collections", collections)

		c.Next()
	}
}

func IsAPIRequest(c *gin.Context) bool {
	isAPI, exists := c.Get("is_api_request")
	if !exists {
		return false
	}
	return isAPI.(bool)
}
func GetSubscription(c *gin.Context) *models.Subscription {
	subscription, exists := c.Get("subscription")
	if !exists {
		return nil
	}
	return subscription.(*models.Subscription)
}
