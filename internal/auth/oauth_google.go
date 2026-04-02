package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"music-curation/internal/user"
	"music-curation/pkg/utils"
)

// GoogleUserInfo represents the response from Google's userinfo API.
type GoogleUserInfo struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

// getGoogleOAuthConfig builds the Google OAuth2 config from environment variables.
func getGoogleOAuthConfig() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),
		Scopes:       []string{"openid", "email", "profile"},
		Endpoint:     google.Endpoint,
	}
}

// GoogleLogin handles GET /api/v1/auth/google — redirects to Google consent screen.
func (h *Handler) GoogleLogin(c *gin.Context) {
	config := getGoogleOAuthConfig()
	url := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

// GoogleCallback handles GET /api/v1/auth/google/callback — exchanges code for token and upserts user.
func (h *Handler) GoogleCallback(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Authorization code is required"})
		return
	}

	config := getGoogleOAuthConfig()

	// Exchange authorization code for access token
	token, err := config.Exchange(c.Request.Context(), code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to exchange token: %v", err)})
		return
	}

	// Fetch user info from Google
	client := config.Client(c.Request.Context(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user info from Google"})
		return
	}
	defer resp.Body.Close()

	var googleUser GoogleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&googleUser); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse Google user info"})
		return
	}

	// Find or create user
	existingUser, err := h.userService.GetUserByGoogleID(googleUser.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	var appUser *user.User

	if existingUser != nil {
		// User already linked with this Google account
		appUser = existingUser
	} else {
		// Check if a user with this email already exists (e.g. signed up with password)
		emailUser, err := h.userService.GetUserByEmail(googleUser.Email)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}

		if emailUser != nil {
			// Link Google ID to existing account
			emailUser.GoogleID = &googleUser.ID
			if err := h.userService.UpdateUser(emailUser); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to link Google account"})
				return
			}
			appUser = emailUser
		} else {
			// Create a new user
			appUser = &user.User{
				Email:    googleUser.Email,
				GoogleID: &googleUser.ID,
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
		"message": "Google authentication successful",
		"data": AuthResponse{
			Token: jwtToken,
			User:  appUser,
		},
	})
}
