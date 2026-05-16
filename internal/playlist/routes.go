package playlist

import (
	"github.com/gin-gonic/gin"

	"music-curation/internal/middleware"
)

func RegisterRoutes(router *gin.RouterGroup, handler *Handler) {
	pl := router.Group("/playlists")
	pl.Use(middleware.AuthMiddleware())
	{
		pl.POST("", handler.Create)
		pl.GET("", handler.List)
		pl.GET("/:id", handler.GetByID)
		pl.PUT("/:id", handler.Rename)
		pl.DELETE("/:id", handler.Delete)
		pl.DELETE("/:id/tracks/:trackId", handler.RemoveTrack)
	}
}
