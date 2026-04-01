package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"music-curation/internal/auth"
	"music-curation/internal/seeder"
	"music-curation/internal/user"
	"music-curation/pkg/database"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  No .env file found, using system environment variables")
	}

	// Connect to PostgreSQL
	db := database.ConnectPostgres()

	// Auto-migrate models
	if err := db.AutoMigrate(&user.User{}); err != nil {
		log.Fatalf("❌ Failed to auto-migrate: %v", err)
	}
	log.Println("✅ Database migration completed")

	// Seed sample users
	seeder.SeedUsers(db)

	// Initialize layers: Repository → Service → Handler
	userRepo := user.NewRepository(db)
	userService := user.NewService(userRepo)
	userHandler := user.NewHandler(userService)
	authHandler := auth.NewHandler(userService)

	// Setup Gin router
	router := gin.Default()

	// API v1 group
	v1 := router.Group("/api/v1")
	{
		auth.RegisterRoutes(v1, authHandler)
		user.RegisterRoutes(v1, userHandler)
	}

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "message": "Music Curation API is running 🎵"})
	})

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("🚀 Server starting on port %s", port)
	if err := router.Run(fmt.Sprintf(":%s", port)); err != nil {
		log.Fatalf("❌ Failed to start server: %v", err)
	}
}
