package spotify

import (
	"music-curation/internal/auth"
	"music-curation/internal/user"
)

// Handler — Spotify handler.
type Handler struct {
	userService  *user.Service
	tokenManager *auth.SpotifyTokenManager
}

// NewHandler — Yeni handler oluşturur.
func NewHandler(userService *user.Service, tokenManager *auth.SpotifyTokenManager) *Handler {
	return &Handler{
		userService:  userService,
		tokenManager: tokenManager,
	}
}
