package server

import (
	"os"
	"strings"
	"time"

	"supergit/inpatient/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func InitializeServer() *gin.Engine {
	r := gin.Default()
	r.SetTrustedProxies(nil)
	r.Use(middleware.RecoveryMiddleware())
	r.Use(middleware.RequestIDMiddleware())
	// r.Use(middleware.RequestLoggerMiddleware())
	r.Use(middleware.SecurityHeadersMiddleware())
	r.Use(middleware.TimeoutMiddleware(30 * time.Second))

	allowedOrigins := os.Getenv("ALLOWED_ORIGINS")
	var origins []string

	if allowedOrigins != "" {
		origins = strings.Split(allowedOrigins, ",")
		for i, origin := range origins {
			origins[i] = strings.TrimSpace(origin)
		}
	} else {
		origins = []string{"*"}
	}

	config := cors.Config{
		AllowOrigins:     origins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"*"},
		ExposeHeaders:    []string{"Content-Length", "X-Request-ID", "X-RateLimit-Limit", "X-RateLimit-Remaining", "X-RateLimit-Reset"},
		AllowCredentials: false,
		MaxAge:           12 * 3600,
	}
	if allowedOrigins != "" && allowedOrigins != "*" {
		config.AllowCredentials = true
	}

	r.Use(cors.New(config))

	return r
}
