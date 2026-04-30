package auth

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"music-curation/pkg/utils"
)

// tokenBlacklist — İptal edilen token'lar.
var (
	tokenBlacklist     = make(map[string]time.Time)
	tokenBlacklistLock sync.RWMutex
)

// BlacklistToken — Token'ı kara listeye ekle.
func BlacklistToken(token string, expiresAt time.Time) {
	tokenBlacklistLock.Lock()
	defer tokenBlacklistLock.Unlock()
	tokenBlacklist[token] = expiresAt
}

// IsTokenBlacklisted — Token iptal edilmiş mi?
func IsTokenBlacklisted(token string) bool {
	tokenBlacklistLock.RLock()
	defer tokenBlacklistLock.RUnlock()
	_, exists := tokenBlacklist[token]
	return exists
}

// CleanupBlacklist — Süresi dolan token'ları temizler.
func CleanupBlacklist() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		tokenBlacklistLock.Lock()
		now := time.Now()
		for token, expiresAt := range tokenBlacklist {
			if now.After(expiresAt) {
				delete(tokenBlacklist, token)
			}
		}
		tokenBlacklistLock.Unlock()
	}
}

// Logout — Token'ı iptal et, çıkış yap.
func (h *Handler) Logout(c *gin.Context) {
	var req LogoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "refresh_token is required"})
		return
	}

	// Token'ı doğrula
	claims, err := utils.ValidateToken(req.RefreshToken)
	if err != nil {
		// Geçersiz olsa bile çıkış başarılı
		c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
		return
	}

	// Refresh token mı?
	if claims.TokenType != "refresh" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid token type, expected refresh token"})
		return
	}

	// İptal et
	BlacklistToken(req.RefreshToken, claims.ExpiresAt.Time)

	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}
