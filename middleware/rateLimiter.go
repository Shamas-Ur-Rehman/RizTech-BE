package middleware

import (
	"fmt"
	"net/http"
	"time"

	"supergit/inpatient/utils"

	"github.com/gin-gonic/gin"
	"github.com/ulule/limiter/v3"
	"github.com/ulule/limiter/v3/drivers/store/memory"
)

func RateLimiterMiddleware() gin.HandlerFunc {
	rate := limiter.Rate{
		Period: 1 * time.Minute,
		Limit:  100,
	}
	store := memory.NewStore()
	instance := limiter.New(store, rate)

	return func(c *gin.Context) {
		if c.Request.Method == "OPTIONS" {
			c.Next()
			return
		}
		limiterCtx, err := instance.Get(c, c.ClientIP())
		if err != nil {
			c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
				Status:  http.StatusInternalServerError,
				Message: "Rate limiter error",
			})
			c.Abort()
			return
		}
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", limiterCtx.Limit))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", limiterCtx.Remaining))
		c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", limiterCtx.Reset))

		if limiterCtx.Reached {
			c.JSON(http.StatusTooManyRequests, utils.ErrorResponse{
				Status:  http.StatusTooManyRequests,
				Message: "Rate limit exceeded. Please try again later.",
				Data: map[string]interface{}{
					"retry_after": limiterCtx.Reset,
				},
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
func StrictRateLimiterMiddleware() gin.HandlerFunc {
	rate := limiter.Rate{
		Period: 1 * time.Minute,
		Limit:  5,
	}
	store := memory.NewStore()
	instance := limiter.New(store, rate)

	return func(c *gin.Context) {
		if c.Request.Method == "OPTIONS" {
			c.Next()
			return
		}
		limiterCtx, err := instance.Get(c, c.ClientIP())
		if err != nil {
			c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
				Status:  http.StatusInternalServerError,
				Message: "Rate limiter error",
			})
			c.Abort()
			return
		}

		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", limiterCtx.Limit))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", limiterCtx.Remaining))
		c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", limiterCtx.Reset))

		if limiterCtx.Reached {
			c.JSON(http.StatusTooManyRequests, utils.ErrorResponse{
				Status:  http.StatusTooManyRequests,
				Message: "Too many attempts. Please try again later.",
				Data: map[string]interface{}{
					"retry_after": limiterCtx.Reset,
				},
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
