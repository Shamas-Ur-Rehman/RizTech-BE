package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

func RequestLoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()
		requestID, _ := c.Get("request_id")
		c.Next()
		latency := time.Since(startTime)
		statusCode := c.Writer.Status()
		log.Printf("[%s] %s %s | Status: %d | Latency: %v | IP: %s | User-Agent: %s",
			requestID,
			c.Request.Method,
			c.Request.URL.Path,
			statusCode,
			latency,
			c.ClientIP(),
			c.Request.UserAgent(),
		)
		if len(c.Errors) > 0 {
			log.Printf("[%s] Errors: %v", requestID, c.Errors.String())
		}
	}
}
