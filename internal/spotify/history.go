package spotify

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// RecentlyPlayedResponse — Son dinlenenler yanıtı.
type RecentlyPlayedResponse struct {
	Items []PlayHistoryItem `json:"items"`
}

// PlayHistoryItem — Tek dinleme kaydı.
type PlayHistoryItem struct {
	Track    Track  `json:"track"`
	PlayedAt string `json:"played_at"`
}

// Track — Şarkı bilgisi.
type Track struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	URI        string   `json:"uri"`
	Artists    []Artist `json:"artists"`
	Album      Album    `json:"album"`
	DurationMs int      `json:"duration_ms"`
}

// Artist — Sanatçı bilgisi.
type Artist struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	URI  string `json:"uri"`
}

// Album — Albüm bilgisi.
type Album struct {
	ID     string   `json:"id"`
	Name   string   `json:"name"`
	Images []Image  `json:"images"`
}

// Image — Görsel bilgisi.
type Image struct {
	URL    string `json:"url"`
	Height int    `json:"height"`
	Width  int    `json:"width"`
}

// GetRecentlyPlayed — Son dinlenen şarkıları getir.
func (h *Handler) GetRecentlyPlayed(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Kullanıcıyı bul
	u, err := h.userService.GetUserByID(userID.(uint))
	if err != nil || u == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Spotify token al
	accessToken, err := h.tokenManager.GetValidToken(u)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Spotify not connected: %v", err)})
		return
	}

	// API'den çek
	limit := c.DefaultQuery("limit", "20")
	var result RecentlyPlayedResponse
	if err := spotifyGet(accessToken, "/me/player/recently-played?limit="+limit, &result); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to fetch history: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Recently played tracks retrieved successfully",
		"data":    result.Items,
		"count":   len(result.Items),
	})
}
