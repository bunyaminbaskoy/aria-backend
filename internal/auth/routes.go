package auth

import (
	"github.com/gin-gonic/gin"

	"music-curation/internal/middleware"
)

// RegisterRoutes — Auth route'larını kaydeder.
func RegisterRoutes(router *gin.RouterGroup, handler *Handler) {
	auth := router.Group("/auth")
	{
		auth.POST("/signup", handler.Signup)
		auth.POST("/login", handler.Login)
		auth.GET("/me", middleware.AuthMiddleware(), handler.Me)

		// Google OAuth giriş
		auth.GET("/google", handler.GoogleLogin)
		auth.GET("/google/callback", handler.GoogleCallback)

		// Spotify OAuth giriş
		auth.GET("/spotify", handler.SpotifyLogin)
		auth.GET("/spotify/callback", handler.SpotifyCallback)

		// Token yönetimi
		auth.POST("/refresh", handler.Refresh)
		auth.POST("/logout", handler.Logout)
	}
}
