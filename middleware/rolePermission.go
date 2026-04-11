package middleware

import (
	"context"
	"net/http"
	"time"

	"supergit/inpatient/utils"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func RolePermissionMiddleware(module, action string, db *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleIDStr, exists := c.Get("role_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, utils.ErrorResponse{Status: http.StatusUnauthorized, Message: "User not authenticated"})
			c.Abort()
			return
		}

		roleID, err := primitive.ObjectIDFromHex(roleIDStr.(string))
		if err != nil {
			c.JSON(http.StatusUnauthorized, utils.ErrorResponse{Status: http.StatusUnauthorized, Message: "Invalid role ID"})
			c.Abort()
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		count, err := db.Collection("permissions").CountDocuments(ctx, bson.M{
			"role_id": roleID,
			"module":  module,
			"action":  action,
		})
		if err != nil || count == 0 {
			c.JSON(http.StatusForbidden, utils.ErrorResponse{Status: http.StatusForbidden, Message: "Insufficient permissions"})
			c.Abort()
			return
		}

		c.Next()
	}
}
