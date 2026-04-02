package auth

import (
	"github.com/gin-gonic/gin"

	"music-curation/internal/middleware"
)

// RegisterRoutes registers all auth routes to the given router group.
func RegisterRoutes(router *gin.RouterGroup, handler *Handler) {
	auth := router.Group("/auth")
	{
		auth.POST("/signup", handler.Signup)
		auth.POST("/login", handler.Login)
		auth.GET("/me", middleware.AuthMiddleware(), handler.Me)

		// Google OAuth
		auth.GET("/google", handler.GoogleLogin)
		auth.GET("/google/callback", handler.GoogleCallback)

		// Spotify OAuth
		auth.GET("/spotify", handler.SpotifyLogin)
		auth.GET("/spotify/callback", handler.SpotifyCallback)
	}
}
