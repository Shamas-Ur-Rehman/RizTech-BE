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
	r.Use(middleware.SecurityHeadersMiddleware())
	r.Use(middleware.TimeoutMiddleware(30 * time.Second))

	allowedOrigins := os.Getenv("ALLOWED_ORIGINS")
	var origins []string
	if allowedOrigins != "" {
		for _, o := range strings.Split(allowedOrigins, ",") {
			origins = append(origins, strings.TrimSpace(o))
		}
	} else {
		origins = []string{"*"}
	}

	corsConfig := cors.Config{
		AllowOrigins:     origins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"*"},
		ExposeHeaders:    []string{"Content-Length", "X-Request-ID"},
		AllowCredentials: allowedOrigins != "" && allowedOrigins != "*",
		MaxAge:           12 * 3600,
	}
	r.Use(cors.New(corsConfig))

	return r
}
