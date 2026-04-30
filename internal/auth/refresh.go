package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"music-curation/pkg/utils"
)

// Refresh — Refresh token ile yeni access token al.
func (h *Handler) Refresh(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "refresh_token is required"})
		return
	}

	// Token'ı doğrula
	claims, err := utils.ValidateToken(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired refresh token"})
		return
	}

	// Refresh token mı?
	if claims.TokenType != "refresh" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token type, expected refresh token"})
		return
	}

	// Kara listede mi?
	if IsTokenBlacklisted(req.RefreshToken) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token has been revoked"})
		return
	}

	// Yeni token üret
	accessToken, err := utils.GenerateAccessToken(claims.UserID, claims.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate access token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Token refreshed successfully",
		"data": gin.H{
			"access_token": accessToken,
		},
	})
}
