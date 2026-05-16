package savedtrack

import (
	"github.com/gin-gonic/gin"

	"music-curation/internal/middleware"
)

func RegisterRoutes(router *gin.RouterGroup, handler *Handler) {
	saved := router.Group("/saved")
	saved.Use(middleware.AuthMiddleware())
	{
		saved.POST("", handler.Save)
		saved.GET("", handler.List)
		saved.DELETE("/:spotify_track_id", handler.Delete)
	}
}
