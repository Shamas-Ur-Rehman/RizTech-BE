package middleware

import (
	"net/http"
	"strings"

	"supergit/inpatient/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func JWTAuth(sqlDB *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == "OPTIONS" {
			c.Next()
			return
		}

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, utils.ErrorResponse{
				Status:  http.StatusUnauthorized,
				Message: "Authorization header is required",
			})
			c.Abort()
			return
		}
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			c.JSON(http.StatusUnauthorized, utils.ErrorResponse{
				Status:  http.StatusUnauthorized,
				Message: "Invalid authorization header format",
			})
			c.Abort()
			return
		}
		claims, err := utils.ValidateJWT(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, utils.ErrorResponse{
				Status:  http.StatusUnauthorized,
				Message: "Invalid token: " + err.Error(),
			})
			c.Abort()
			return
		}
		c.Set("user_id", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("role_id", claims.RoleID)
		c.Set("db_name", claims.DbName)
		c.Set("sql_db", sqlDB)

		c.Next()
	}
}

func RoleAuth(requiredRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleID, exists := c.Get("role_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, utils.ErrorResponse{
				Status:  http.StatusUnauthorized,
				Message: "User role not found",
			})
			c.Abort()
			return
		}

		userRole := roleID.(string)
		hasRequiredRole := false
		for _, role := range requiredRoles {
			if userRole == role {
				hasRequiredRole = true
				break
			}
		}

		if !hasRequiredRole {
			c.JSON(http.StatusForbidden, utils.ErrorResponse{
				Status:  http.StatusForbidden,
				Message: "Insufficient permissions",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
