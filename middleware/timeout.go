package middleware

import (
	"context"
	"net/http"
	"time"

	"supergit/inpatient/utils"

	"github.com/gin-gonic/gin"
)

func TimeoutMiddleware(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()
		c.Request = c.Request.WithContext(ctx)
		finished := make(chan struct{})

		go func() {
			c.Next()
			finished <- struct{}{}
		}()
		select {
		case <-finished:
			return
		case <-ctx.Done():
			c.JSON(http.StatusRequestTimeout, utils.ErrorResponse{
				Status:  http.StatusRequestTimeout,
				Message: "Request timeout. Please try again.",
			})
			c.Abort()
		}
	}
}
