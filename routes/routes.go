package routes

import (
	"supergit/inpatient/controllers"
	"supergit/inpatient/middleware"
	"supergit/inpatient/server"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

func SetupRouter(db *mongo.Database) *gin.Engine {
	r := server.InitializeServer()

	api := r.Group("/api/v1")

	// ── Public auth ─────────────────────────────────────────────────────────
	auth := api.Group("/auth")
	{
		auth.POST("/register", func(c *gin.Context) { controllers.Register(c, db) })
		auth.POST("/login", func(c *gin.Context) { controllers.Login(c, db) })
		auth.POST("/forgot-password", func(c *gin.Context) { controllers.ForgotPassword(c, db) })
		auth.POST("/reset-password", func(c *gin.Context) { controllers.ResetPassword(c, db) })
	}

	// ── Public products ──────────────────────────────────────────────────────
	api.GET("/products", func(c *gin.Context) { controllers.ListProducts(c, db) })
	api.GET("/products/:id", func(c *gin.Context) { controllers.GetProduct(c, db) })

	// ── Public shipping settings (checkout needs this) ───────────────────────
	api.GET("/shipping-settings", func(c *gin.Context) { controllers.GetShippingSettings(c, db) })

	// ── Authenticated user routes ────────────────────────────────────────────
	user := api.Group("/")
	user.Use(middleware.JWTAuth())
	{
		user.GET("/me", func(c *gin.Context) { controllers.GetMe(c, db) })
		user.POST("/orders", func(c *gin.Context) { controllers.PlaceOrder(c, db) })
		user.GET("/my-orders", func(c *gin.Context) { controllers.GetMyOrders(c, db) })
	}

	// ── Admin-only routes ────────────────────────────────────────────────────
	admin := api.Group("/admin")
	admin.Use(middleware.JWTAuth(), middleware.AdminOnly())
	{
		admin.GET("/me", func(c *gin.Context) { controllers.GetMe(c, db) })

		// Stats
		admin.GET("/stats", func(c *gin.Context) { controllers.GetStats(c, db) })

		// Change password
		admin.PUT("/change-password", func(c *gin.Context) { controllers.ChangePassword(c, db) })

		// Image upload (Cloudinary)
		admin.POST("/upload", controllers.UploadImage)

		// Products CRUD
		admin.POST("/products", func(c *gin.Context) { controllers.CreateProduct(c, db) })
		admin.PUT("/products/:id", func(c *gin.Context) { controllers.UpdateProduct(c, db) })
		admin.DELETE("/products/:id", func(c *gin.Context) { controllers.DeleteProduct(c, db) })

		// Orders
		admin.GET("/orders", func(c *gin.Context) { controllers.ListOrders(c, db) })
		admin.PATCH("/orders/:id/status", func(c *gin.Context) { controllers.UpdateOrderStatus(c, db) })
		admin.PUT("/orders/:id/tracking", func(c *gin.Context) { controllers.AddTracking(c, db) })

		// Shipping settings
		admin.PUT("/shipping-settings", func(c *gin.Context) { controllers.SaveShippingSettings(c, db) })

		// Users management
		admin.GET("/users", func(c *gin.Context) { controllers.ListUsers(c, db) })
		admin.PATCH("/users/:id/status", func(c *gin.Context) { controllers.ToggleUserStatus(c, db) })
	}

	return r
}
