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

// GoogleUserInfo — Google kullanıcı bilgisi.
type GoogleUserInfo struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

// getGoogleOAuthConfig — Google OAuth2 ayarları.
func getGoogleOAuthConfig() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),
		Scopes:       []string{"openid", "email", "profile"},
		Endpoint:     google.Endpoint,
	}
}

// GoogleLogin — Google giriş sayfasına yönlendir.
func (h *Handler) GoogleLogin(c *gin.Context) {
	config := getGoogleOAuthConfig()
	url := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

// GoogleCallback — Google'dan dönen kodu işle.
func (h *Handler) GoogleCallback(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Authorization code is required"})
		return
	}

	config := getGoogleOAuthConfig()

	// Kodu token'a çevir
	token, err := config.Exchange(c.Request.Context(), code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to exchange token: %v", err)})
		return
	}

	// Kullanıcı bilgisini çek
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

	// Bul veya oluştur
	existingUser, err := h.userService.GetUserByGoogleID(googleUser.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	var appUser *user.User

	if existingUser != nil {
		// Zaten eşleşmiş
		appUser = existingUser
	} else {
		// Email ile kayıtlı mı?
		emailUser, err := h.userService.GetUserByEmail(googleUser.Email)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}

		if emailUser != nil {
			// Google'u hesaba bağla
			emailUser.GoogleID = &googleUser.ID
			if err := h.userService.UpdateUser(emailUser); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to link Google account"})
				return
			}
			appUser = emailUser
		} else {
			// Yeni kayıt
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

	// Token çifti üret
	tokenPair, err := utils.GenerateTokenPair(appUser.ID, appUser.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:5173"
	}
	redirectURL := fmt.Sprintf("%s/auth/callback?access_token=%s&refresh_token=%s",
		frontendURL, tokenPair.AccessToken, tokenPair.RefreshToken)
	c.Redirect(http.StatusTemporaryRedirect, redirectURL)
}
