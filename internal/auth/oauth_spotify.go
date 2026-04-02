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

// SpotifyUserInfo represents the response from Spotify's /me API.
type SpotifyUserInfo struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

// spotifyEndpoint is the OAuth2 endpoint for Spotify.
var spotifyEndpoint = oauth2.Endpoint{
	AuthURL:  "https://accounts.spotify.com/authorize",
	TokenURL: "https://accounts.spotify.com/api/token",
}

// getSpotifyOAuthConfig builds the Spotify OAuth2 config from environment variables.
func getSpotifyOAuthConfig() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     os.Getenv("SPOTIFY_CLIENT_ID"),
		ClientSecret: os.Getenv("SPOTIFY_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("SPOTIFY_REDIRECT_URL"),
		Scopes:       []string{"user-read-email", "user-read-private", "playlist-modify-public", "playlist-modify-private"},
		Endpoint:     spotifyEndpoint,
	}
}

// SpotifyLogin handles GET /api/v1/auth/spotify — redirects to Spotify authorize page.
func (h *Handler) SpotifyLogin(c *gin.Context) {
	config := getSpotifyOAuthConfig()
	url := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

// SpotifyCallback handles GET /api/v1/auth/spotify/callback — exchanges code for token and upserts user.
func (h *Handler) SpotifyCallback(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Authorization code is required"})
		return
	}

	config := getSpotifyOAuthConfig()

	// Exchange authorization code for access token
	token, err := config.Exchange(c.Request.Context(), code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to exchange token: %v", err)})
		return
	}

	// Fetch user info from Spotify
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

	// Find or create user
	existingUser, err := h.userService.GetUserBySpotifyID(spotifyUser.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	var appUser *user.User

	if existingUser != nil {
		// User already linked — update Spotify tokens
		existingUser.SpotifyAccessToken = &token.AccessToken
		existingUser.SpotifyRefreshToken = &token.RefreshToken
		if err := h.userService.UpdateUser(existingUser); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update Spotify tokens"})
			return
		}
		appUser = existingUser
	} else {
		// Check if a user with this email already exists
		emailUser, err := h.userService.GetUserByEmail(spotifyUser.Email)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}

		if emailUser != nil {
			// Link Spotify to existing account
			emailUser.SpotifyID = &spotifyUser.ID
			emailUser.SpotifyAccessToken = &token.AccessToken
			emailUser.SpotifyRefreshToken = &token.RefreshToken
			if err := h.userService.UpdateUser(emailUser); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to link Spotify account"})
				return
			}
			appUser = emailUser
		} else {
			// Create a new user
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

	// Generate JWT
	jwtToken, err := utils.GenerateToken(appUser.ID, appUser.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Spotify authentication successful",
		"data": AuthResponse{
			Token: jwtToken,
			User:  appUser,
		},
	})
}
