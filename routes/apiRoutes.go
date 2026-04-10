package routes

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"

	"supergit/inpatient/controllers"
	"supergit/inpatient/middleware"
)

func SetupAPIRoutes(r *gin.Engine, sqlDB *gorm.DB, mongoClient *mongo.Client) {
	api := r.Group("/api/v1")
	api.Use(middleware.APIKeyMiddleware(sqlDB, mongoClient))
	api.Use(middleware.RateLimiterMiddleware())

	patientAPI := api.Group("/patients")
	{
		patientAPI.POST("", func(c *gin.Context) {controllers.CreatePatient(c, mongoClient)})
	}

	// Health check endpoint for API
	api.GET("/health", func(c *gin.Context) {
		user := middleware.GetUser(c)
		subscription := middleware.GetSubscription(c)
		c.JSON(200, gin.H{
			"status":      "ok",
			"system_user": user.FullName,
			"business_id": user.BusinessId,
			"branch_id":   user.BranchId,
			"subscription_id": subscription.ID,
		})
	})
}
