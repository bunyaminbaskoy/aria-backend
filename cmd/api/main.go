package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"music-curation/internal/auth"
	"music-curation/internal/middleware"
	"music-curation/internal/seeder"
	"music-curation/internal/spotify"
	"music-curation/internal/user"
	"music-curation/pkg/database"
)

func main() {
	// .env dosyasını yükle
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  No .env file found, using system environment variables")
	}

	// Veritabanına bağlan
	db := database.ConnectPostgres()

	// Tabloları oluştur
	if err := db.AutoMigrate(&user.User{}); err != nil {
		log.Fatalf("❌ Failed to auto-migrate: %v", err)
	}
	log.Println("✅ Database migration completed")

	// Örnek veriler
	seeder.SeedUsers(db)

	// Katmanları başlat
	userRepo := user.NewRepository(db)
	userService := user.NewService(userRepo)
	userHandler := user.NewHandler(userService)
	authHandler := auth.NewHandler(userService)
	spotifyTokenManager := auth.NewSpotifyTokenManager(userService)
	spotifyHandler := spotify.NewHandler(userService, spotifyTokenManager)

	// Token temizliği başlat
	go auth.CleanupBlacklist()

	// Router ve middleware'ler
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(middleware.RecoveryHandler())
	router.Use(middleware.ErrorHandler())
	router.Use(middleware.RateLimitMiddleware())

	// Route'lar
	v1 := router.Group("/api/v1")
	{
		auth.RegisterRoutes(v1, authHandler)
		user.RegisterRoutes(v1, userHandler)
		spotify.RegisterRoutes(v1, spotifyHandler)
	}

	// Sağlık kontrolü
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "message": "Music Curation API is running 🎵"})
	})

	// Sunucuyu başlat
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("🚀 Server starting on port %s", port)
	if err := router.Run(fmt.Sprintf(":%s", port)); err != nil {
		log.Fatalf("❌ Failed to start server: %v", err)
	}
}
