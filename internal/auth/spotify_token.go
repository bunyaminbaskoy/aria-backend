package auth

import (
	"context"
	"fmt"
	"os"
	"time"

	"golang.org/x/oauth2"

	"music-curation/internal/user"
)

// SpotifyTokenManager — Spotify token yenileme yöneticisi.
type SpotifyTokenManager struct {
	userService *user.Service
}

// NewSpotifyTokenManager — Yeni yönetici oluşturur.
func NewSpotifyTokenManager(userService *user.Service) *SpotifyTokenManager {
	return &SpotifyTokenManager{userService: userService}
}

// GetValidToken — Geçerli token döndür, gerekirse yenile.
func (m *SpotifyTokenManager) GetValidToken(u *user.User) (string, error) {
	if u.SpotifyAccessToken == nil || u.SpotifyRefreshToken == nil {
		return "", fmt.Errorf("user has no Spotify tokens")
	}

	// Mevcut token'larla oauth2.Token oluştur
	tok := &oauth2.Token{
		AccessToken:  *u.SpotifyAccessToken,
		RefreshToken: *u.SpotifyRefreshToken,
		// Yenilemeyi zorla
		Expiry: time.Now().Add(-1 * time.Minute),
	}

	config := &oauth2.Config{
		ClientID:     os.Getenv("SPOTIFY_CLIENT_ID"),
		ClientSecret: os.Getenv("SPOTIFY_CLIENT_SECRET"),
		Endpoint:     spotifyEndpoint,
	}

	// Otomatik yenile
	tokenSource := config.TokenSource(context.Background(), tok)
	newToken, err := tokenSource.Token()
	if err != nil {
		return "", fmt.Errorf("failed to refresh Spotify token: %w", err)
	}

	// Değiştiyse DB'ye kaydet
	if newToken.AccessToken != *u.SpotifyAccessToken {
		u.SpotifyAccessToken = &newToken.AccessToken
		if newToken.RefreshToken != "" {
			u.SpotifyRefreshToken = &newToken.RefreshToken
		}
		if err := m.userService.UpdateUser(u); err != nil {
			return "", fmt.Errorf("failed to save refreshed Spotify token: %w", err)
		}
	}

	return newToken.AccessToken, nil
}
