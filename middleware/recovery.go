package middleware

import (
	"fmt"
	"log"
	"net/http"
	"runtime/debug"

	"supergit/inpatient/utils"

	"github.com/gin-gonic/gin"
)

func RecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				requestID, _ := c.Get("request_id")
				log.Printf("[PANIC RECOVERED] Request ID: %v | Error: %v\n%s",
					requestID,
					err,
					debug.Stack(),
				)
				c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
					Status:  http.StatusInternalServerError,
					Message: "Internal server error. Please try again later.",
					Data: map[string]interface{}{
						"request_id": fmt.Sprintf("%v", requestID),
					},
				})

				c.Abort()
			}
		}()

		c.Next()
	}
}
