package middleware

import (
	"net/http"
	"os"

	"supergit/inpatient/utils"

	"github.com/gin-gonic/gin"
)

func VersionCheckMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == "OPTIONS" {
			c.Next()
			return
		}

		clientVersion := c.GetHeader("version")
		expectedVersion := os.Getenv("VERSION")
		if expectedVersion == "" {
			c.Next()
			return
		}
		if clientVersion == "" {
			c.JSON(http.StatusBadRequest, utils.ErrorResponse{
				Status:  http.StatusBadRequest,
				Message: "Version header is required",
			})
			c.Abort()
			return
		}
		if clientVersion != expectedVersion {
			c.JSON(http.StatusUpgradeRequired, utils.ErrorResponse{
				Status:  http.StatusUpgradeRequired,
				Message: "Version mismatch. Please update your application.",
				Data: map[string]interface{}{
					"expected_version": expectedVersion,
					"client_version":   clientVersion,
				},
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
