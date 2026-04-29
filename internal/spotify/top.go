package spotify

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// TopTracksResponse — En çok dinlenen şarkılar yanıtı.
type TopTracksResponse struct {
	Items []Track `json:"items"`
	Total int     `json:"total"`
}

// TopArtistsResponse — En çok dinlenen sanatçılar yanıtı.
type TopArtistsResponse struct {
	Items []FullArtist `json:"items"`
	Total int          `json:"total"`
}

// FullArtist — Detaylı sanatçı bilgisi.
type FullArtist struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	URI        string   `json:"uri"`
	Genres     []string `json:"genres"`
	Images     []Image  `json:"images"`
	Popularity int      `json:"popularity"`
}

// GetTopTracks — En çok dinlenen şarkıları getir.
func (h *Handler) GetTopTracks(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	u, err := h.userService.GetUserByID(userID.(uint))
	if err != nil || u == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	accessToken, err := h.tokenManager.GetValidToken(u)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Spotify not connected: %v", err)})
		return
	}

	timeRange := c.DefaultQuery("time_range", "medium_term")
	limit := c.DefaultQuery("limit", "20")

	var result TopTracksResponse
	endpoint := fmt.Sprintf("/me/top/tracks?time_range=%s&limit=%s", timeRange, limit)
	if err := spotifyGet(accessToken, endpoint, &result); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to fetch top tracks: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Top tracks retrieved successfully",
		"data":    result.Items,
		"count":   len(result.Items),
	})
}

// GetTopArtists — En çok dinlenen sanatçıları getir.
func (h *Handler) GetTopArtists(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	u, err := h.userService.GetUserByID(userID.(uint))
	if err != nil || u == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	accessToken, err := h.tokenManager.GetValidToken(u)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Spotify not connected: %v", err)})
		return
	}

	timeRange := c.DefaultQuery("time_range", "medium_term")
	limit := c.DefaultQuery("limit", "20")

	var result TopArtistsResponse
	endpoint := fmt.Sprintf("/me/top/artists?time_range=%s&limit=%s", timeRange, limit)
	if err := spotifyGet(accessToken, endpoint, &result); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to fetch top artists: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Top artists retrieved successfully",
		"data":    result.Items,
		"count":   len(result.Items),
	})
}
