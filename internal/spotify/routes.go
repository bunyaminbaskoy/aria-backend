package spotify

import (
	"github.com/gin-gonic/gin"

	"music-curation/internal/middleware"
)

// RegisterRoutes — Spotify route'larını kaydeder.
func RegisterRoutes(router *gin.RouterGroup, handler *Handler) {
	spotify := router.Group("/spotify")
	spotify.Use(middleware.AuthMiddleware())
	{
		// Dinleme geçmişi
		spotify.GET("/history", handler.GetRecentlyPlayed)

		// En çok dinlenen şarkılar ve sanatçılar
		spotify.GET("/top/tracks", handler.GetTopTracks)
		spotify.GET("/top/artists", handler.GetTopArtists)

		// Playlist oluşturma
		spotify.POST("/playlist", handler.CreatePlaylist)
	}
}
