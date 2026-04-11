package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"

	"supergit/inpatient/config"
	"supergit/inpatient/routes"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	mongoClient := config.ConnectMongoDB()
	defer mongoClient.Disconnect(nil)

	r := routes.SetupRouter(config.MongoDB)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on :%s\n", port)
	if err := r.Run(fmt.Sprintf(":%s", port)); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
