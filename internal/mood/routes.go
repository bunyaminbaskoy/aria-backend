package mood

import (
	"github.com/gin-gonic/gin"

	"music-curation/internal/middleware"
)

// RegisterRoutes, Mood modülünün tüm HTTP route'larını verilen router
// grubuna kaydeder. Tüm endpoint'ler AuthMiddleware ile korunur —
// kullanıcı yalnızca kendi ruh hali kayıtlarına erişebilir.
func RegisterRoutes(router *gin.RouterGroup, handler *Handler) {
	moods := router.Group("/moods")
	moods.Use(middleware.AuthMiddleware())
	{
		moods.POST("", handler.Create)
		moods.GET("", handler.List)
		moods.GET("/:id", handler.GetByID)
	}
}
