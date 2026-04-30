package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"

	"music-curation/internal/user"
	"music-curation/pkg/utils"
)

// SpotifyUserInfo — Spotify kullanıcı bilgisi.
type SpotifyUserInfo struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

// spotifyEndpoint — Spotify OAuth2 URL'leri.
var spotifyEndpoint = oauth2.Endpoint{
	AuthURL:  "https://accounts.spotify.com/authorize",
	TokenURL: "https://accounts.spotify.com/api/token",
}

// getSpotifyOAuthConfig — Spotify OAuth2 ayarları.
func getSpotifyOAuthConfig() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     os.Getenv("SPOTIFY_CLIENT_ID"),
		ClientSecret: os.Getenv("SPOTIFY_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("SPOTIFY_REDIRECT_URL"),
		Scopes:       []string{"user-read-email", "user-read-private", "playlist-modify-public", "playlist-modify-private", "user-read-recently-played", "user-top-read"},
		Endpoint:     spotifyEndpoint,
	}
}

// SpotifyLogin — Spotify giriş sayfasına yönlendir.
func (h *Handler) SpotifyLogin(c *gin.Context) {
	config := getSpotifyOAuthConfig()
	url := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

// SpotifyCallback — Spotify'dan dönen kodu işle.
func (h *Handler) SpotifyCallback(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Authorization code is required"})
		return
	}

	config := getSpotifyOAuthConfig()

	// Kodu token'a çevir
	token, err := config.Exchange(c.Request.Context(), code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to exchange token: %v", err)})
		return
	}

	// Kullanıcı bilgisini çek
	client := config.Client(c.Request.Context(), token)
	resp, err := client.Get("https://api.spotify.com/v1/me")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user info from Spotify"})
		return
	}
	defer resp.Body.Close()

	var spotifyUser SpotifyUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&spotifyUser); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse Spotify user info"})
		return
	}

	// Bul veya oluştur
	existingUser, err := h.userService.GetUserBySpotifyID(spotifyUser.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	var appUser *user.User

	if existingUser != nil {
		// Zaten eşleşmiş — token'ları güncelle
		existingUser.SpotifyAccessToken = &token.AccessToken
		existingUser.SpotifyRefreshToken = &token.RefreshToken
		if err := h.userService.UpdateUser(existingUser); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update Spotify tokens"})
			return
		}
		appUser = existingUser
	} else {
		// Email ile kayıtlı mı?
		emailUser, err := h.userService.GetUserByEmail(spotifyUser.Email)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}

		if emailUser != nil {
			// Spotify'u hesaba bağla
			emailUser.SpotifyID = &spotifyUser.ID
			emailUser.SpotifyAccessToken = &token.AccessToken
			emailUser.SpotifyRefreshToken = &token.RefreshToken
			if err := h.userService.UpdateUser(emailUser); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to link Spotify account"})
				return
			}
			appUser = emailUser
		} else {
			// Yeni kayıt
			appUser = &user.User{
				Email:               spotifyUser.Email,
				SpotifyID:           &spotifyUser.ID,
				SpotifyAccessToken:  &token.AccessToken,
				SpotifyRefreshToken: &token.RefreshToken,
			}
			if err := h.userService.CreateUser(appUser); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
				return
			}
		}
	}

	// Token çifti üret
	tokenPair, err := utils.GenerateTokenPair(appUser.ID, appUser.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Spotify authentication successful",
		"data": AuthResponse{
			AccessToken:  tokenPair.AccessToken,
			RefreshToken: tokenPair.RefreshToken,
			User:         appUser,
		},
	})
}
