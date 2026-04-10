package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/gorm"

	"supergit/inpatient/config"
	"supergit/inpatient/middleware"
	"supergit/inpatient/routes"
	"supergit/inpatient/utils"
)

var sqlDB *gorm.DB

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	sqlDB = config.ConnectSQLDB()
	if sqlDB == nil {
		log.Fatal("SQL DB connection failed")
	}
	log.Println("DB connected Successfully")
	log.Println("Initializing permission cache...")
	if err := middleware.RefreshPermissionCache(sqlDB); err != nil {
		log.Printf("Warning: Failed to initialize permission cache: %v", err)
	} else {
		log.Println("Permission cache initialized successfully")
	}

	log.Println("Initializing MinIO client...")
	_ = utils.GetMinioClient()
	log.Println("MinIO client initialized successfully")

	// if err := utils.SeedRBAC(sqlDB); err != nil {
	// 	log.Printf("Warning: Seeder failed: %v", err)
	// }

	mongoClient := config.ConnectMongoDB()
	// r := gin.Default()
	r := routes.SetupRouter(sqlDB, mongoClient)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("starting server on :%s\n", port)
	if err := r.Run(fmt.Sprintf(":%s", port)); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
