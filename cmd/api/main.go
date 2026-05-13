package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"music-curation/internal/auth"
	"music-curation/internal/interaction"
	"music-curation/internal/middleware"
	"music-curation/internal/mood"
	"music-curation/internal/pipeline"
	"music-curation/internal/recommendation"
	"music-curation/internal/seeder"
	"music-curation/internal/spotify"
	"music-curation/internal/user"
	"music-curation/pkg/aiclient"
	"music-curation/pkg/cache"
	"music-curation/pkg/database"
)

// main, tüm uygulamanın TEK giriş noktasıdır. Modular Monolith mimarisi
// gereği user, auth, mood, recommendation ve pipeline modülleri aynı
// süreç içinde çalışır ve birbirleriyle direkt Go fonksiyon çağrısı
// üzerinden konuşur — ağ üzerinden DEĞİL. Tüm dependency injection
// burada, tek dosyada yapılır; bu da modüller arası bağımlılıkları
// görünür ve denetlenebilir kılar.
func main() {
	// .env dosyasını yükle (yoksa sistem env'i kullanılır).
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  No .env file found, using system environment variables")
	}

	// --- Altyapı bağlantıları (Postgres + Redis) ---
	db := database.ConnectPostgres()
	rdb := cache.ConnectRedis()

	// AutoMigrate — modeller dependency sırasına göre listelenir.
	// Önce parent (User), sonra child'lar (Mood → User, Recommendation
	// → User+Mood, RecommendedTrack → Recommendation).
	if err := db.AutoMigrate(
		&user.User{},
		&mood.Mood{},
		&recommendation.Recommendation{},
		&recommendation.RecommendedTrack{},
		&interaction.TrackInteraction{},
	); err != nil {
		log.Fatalf("❌ Failed to auto-migrate: %v", err)
	}
	log.Println("✅ Database migration completed")

	// Örnek kullanıcıları seed et (idempotent).
	seeder.SeedUsers(db)

	// --- Dış servisler (AI/RAG client) ---
	// Tek bir Client örneği uygulama yaşam döngüsü boyunca paylaşılır;
	// http.Client'in kendi connection pool'u sayesinde performanslıdır.
	aiClient := aiclient.NewClient()

	// --- Modül wiring (Repository → Service → Handler) ---
	userRepo := user.NewRepository(db)
	userService := user.NewService(userRepo)
	userHandler := user.NewHandler(userService)
	authHandler := auth.NewHandler(userService)
	spotifyTokenManager := auth.NewSpotifyTokenManager(userService)
	spotifyHandler := spotify.NewHandler(userService, spotifyTokenManager)

	moodRepo := mood.NewRepository(db)
	moodService := mood.NewService(moodRepo)
	moodHandler := mood.NewHandler(moodService)

	recRepo := recommendation.NewRepository(db)
	recService := recommendation.NewService(recRepo)
	recHandler := recommendation.NewHandler(recService)

	interactionRepo := interaction.NewRepository(db)
	interactionService := interaction.NewService(interactionRepo)
	interactionHandler := interaction.NewHandler(interactionService)

	// Orchestrator — moodService, recService, aiClient ve
	// interactionService'i direkt Go çağrıları ile tüketir.
	pipelineService := pipeline.NewService(moodService, recService, aiClient, interactionService)
	pipelineHandler := pipeline.NewHandler(pipelineService)

	// Token temizliği başlat.
	go auth.CleanupBlacklist()

	// --- HTTP router ---
	router := gin.Default()
	router.Use(middleware.ErrorHandler())

	// /api/v1 grubu altında tüm modül route'ları kaydedilir.
	v1 := router.Group("/api/v1")
	{
		auth.RegisterRoutes(v1, authHandler)
		user.RegisterRoutes(v1, userHandler)
		spotify.RegisterRoutes(v1, spotifyHandler)
		mood.RegisterRoutes(v1, moodHandler)
		recommendation.RegisterRoutes(v1, recHandler)
		interaction.RegisterRoutes(v1, interactionHandler)

		// Orchestrator endpoint'i için ek middleware: rate limiting.
		// Sadece bu endpoint'e uygulanıyor (kullanıcı başına dakikada
		// 5 playlist üretimi). Diğer modüller etkilenmez.
		generateLimiter := middleware.RateLimitMiddleware(rdb, "generate", 5, time.Minute)
		pipeline.RegisterRoutes(v1, pipelineHandler, generateLimiter)
	}

	// Sağlık kontrolü (health check) — load balancer / Docker için.
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "message": "Music Curation API is running 🎵"})
	})

	// Sunucuyu başlat.
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("🚀 Server starting on port %s", port)
	if err := router.Run(fmt.Sprintf(":%s", port)); err != nil {
		log.Fatalf("❌ Failed to start server: %v", err)
	}
}
