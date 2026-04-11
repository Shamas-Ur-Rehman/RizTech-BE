package middleware

import (
	"net/http"
	"strings"

	"supergit/inpatient/models"
	"supergit/inpatient/utils"

	"github.com/gin-gonic/gin"
)

func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == "OPTIONS" {
			c.Next()
			return
		}
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, utils.ErrorResponse{Status: http.StatusUnauthorized, Message: "Authorization header is required"})
			c.Abort()
			return
		}
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			c.JSON(http.StatusUnauthorized, utils.ErrorResponse{Status: http.StatusUnauthorized, Message: "Invalid authorization header format"})
			c.Abort()
			return
		}
		claims, err := utils.ValidateJWT(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, utils.ErrorResponse{Status: http.StatusUnauthorized, Message: "Invalid token: " + err.Error()})
			c.Abort()
			return
		}
		c.Set("user_id", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("role_id", claims.RoleID)
		c.Set("role_name", claims.RoleName)
		c.Next()
	}
}

// AdminOnly allows only users with the admin role
func AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		roleName, _ := c.Get("role_name")
		if roleName != models.RoleAdmin {
			c.JSON(http.StatusForbidden, utils.ErrorResponse{Status: http.StatusForbidden, Message: "Admin access required"})
			c.Abort()
			return
		}
		c.Next()
	}
}
