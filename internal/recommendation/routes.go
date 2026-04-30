package recommendation

import (
	"github.com/gin-gonic/gin"

	"music-curation/internal/middleware"
)

// RegisterRoutes, Recommendation modülünün HTTP route'larını verilen
// router grubuna kaydeder. Tüm endpoint'ler AuthMiddleware ile korunur.
//
// Yazma route'u (POST) bilinçli olarak yoktur — Recommendation kayıtları
// yalnızca orchestrator pipeline'ı üzerinden, Service.CreateFromAI ile
// oluşturulur. Bu, AI sonuçlarının dışarıdan manipüle edilmesini engeller.
func RegisterRoutes(router *gin.RouterGroup, handler *Handler) {
	recs := router.Group("/recommendations")
	recs.Use(middleware.AuthMiddleware())
	{
		recs.GET("", handler.List)
		recs.GET("/:id", handler.GetByID)
	}

	// Mood-bazlı sorgu için ek route: GET /api/v1/moods/:id/recommendations.
	// Bu route'u burada tanımlamak, Mood modülünün Recommendation modülüne
	// bağımlı olmasını engeller (modüler monolith içinde tek yönlü bağımlılık).
	moodScoped := router.Group("/moods")
	moodScoped.Use(middleware.AuthMiddleware())
	{
		moodScoped.GET("/:id/recommendations", handler.GetByMoodID)
	}
}
