package interaction

import (
	"github.com/gin-gonic/gin"

	"music-curation/internal/middleware"
)

// RegisterRoutes, Interaction modülünün HTTP route'larını verilen
// router grubuna kaydeder. Tüm endpoint'ler AuthMiddleware ile korunur.
func RegisterRoutes(router *gin.RouterGroup, handler *Handler) {
	interactions := router.Group("/interactions")
	interactions.Use(middleware.AuthMiddleware())
	{
		interactions.POST("", handler.Create)
		interactions.GET("", handler.List)
		interactions.DELETE("/:spotify_track_id", handler.Delete)
	}
}
