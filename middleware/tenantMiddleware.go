package middleware

import (
	"net/http"

	"supergit/inpatient/config"
	"supergit/inpatient/models"
	"supergit/inpatient/utils"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"
)

func TenantMiddleware(sqlDB *gorm.DB, mongoClient *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDValue, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, utils.ErrorResponse{
				Status:  http.StatusUnauthorized,
				Message: "User not found in token",
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
		var user models.User
		if err := sqlDB.First(&user, userID).Error; err != nil {
			c.JSON(http.StatusNotFound, utils.ErrorResponse{
				Status:  http.StatusNotFound,
				Message: "User not found",
			})
			c.Abort()
			return
		}
		var business models.Business
		if err := sqlDB.Where("id = ?", user.BusinessId).First(&business).Error; err != nil {
			c.JSON(http.StatusNotFound, utils.ErrorResponse{
				Status:  http.StatusNotFound,
				Message: "Business not found",
			})
			c.Abort()
			return
		}
		dbName := business.InPatientDB
		if dbName == "" {
			dbName = "inpatient_main_db"
		}

		c.Set("business_id", user.BusinessId)
		c.Set("branch_id", user.BranchId)
		c.Set("tenant_db", dbName)
		c.Set("business", business)
		c.Set("user", &user)

		collections := config.GetCollections(mongoClient, dbName)
		c.Set("collections", collections)

		c.Next()
	}
}

func GetTenantDB(c *gin.Context) string {
	dbName, exists := c.Get("tenant_db")
	if !exists {
		return "inpatient_main_db"
	}
	return dbName.(string)
}
func GetCollections(c *gin.Context) *config.Collections {
	collections, exists := c.Get("collections")
	if !exists {
		return nil
	}
	return collections.(*config.Collections)
}
func GetBusinessID(c *gin.Context) uint {
	businessID, exists := c.Get("business_id")
	if !exists {
		return 0
	}
	return businessID.(uint)
}
func GetBranchID(c *gin.Context) uint {
	branchID, exists := c.Get("branch_id")
	if !exists {
		return 0
	}
	return branchID.(uint)
}
func GetUser(c *gin.Context) *models.User {
	user, exists := c.Get("user")
	if !exists {
		return nil
	}
	return user.(*models.User)
}
func GetBusiness(c *gin.Context) *models.Business {
	business, exists := c.Get("business")
	if !exists {
		return nil
	}
	return business.(*models.Business)
}
