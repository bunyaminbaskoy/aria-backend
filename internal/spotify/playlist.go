package spotify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// CreatePlaylistRequest — Playlist oluşturma isteği.
type CreatePlaylistRequest struct {
	Name        string   `json:"name" binding:"required"`
	Description string   `json:"description"`
	Public      bool     `json:"public"`
	TrackURIs   []string `json:"track_uris" binding:"required,min=1"`
}

// SpotifyCreatePlaylistRequest — Spotify'a gönderilen playlist verisi.
type SpotifyCreatePlaylistRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Public      bool   `json:"public"`
}

// SpotifyPlaylistResponse — Oluşturulan playlist yanıtı.
type SpotifyPlaylistResponse struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	ExternalURLs struct {
		Spotify string `json:"spotify"`
	} `json:"external_urls"`
	URI string `json:"uri"`
}

// SpotifyAddTracksRequest — Şarkı ekleme verisi.
type SpotifyAddTracksRequest struct {
	URIs []string `json:"uris"`
}

// CreatePlaylist — Playlist oluştur ve şarkı ekle.
func (h *Handler) CreatePlaylist(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req CreatePlaylistRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	u, err := h.userService.GetUserByID(userID.(uint))
	if err != nil || u == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	if u.SpotifyID == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Spotify account not linked"})
		return
	}

	accessToken, err := h.tokenManager.GetValidToken(u)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Spotify not connected: %v", err)})
		return
	}

	// Adım 1: Spotify'da playlist oluştur
	createBody, _ := json.Marshal(SpotifyCreatePlaylistRequest{
		Name:        req.Name,
		Description: req.Description,
		Public:      req.Public,
	})

	var playlist SpotifyPlaylistResponse
	endpoint := fmt.Sprintf("/users/%s/playlists", *u.SpotifyID)
	if err := spotifyPost(accessToken, endpoint, bytes.NewReader(createBody), &playlist); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to create playlist: %v", err)})
		return
	}

	// Adım 2: Playlist'e şarkıları ekle
	addBody, _ := json.Marshal(SpotifyAddTracksRequest{
		URIs: req.TrackURIs,
	})

	addEndpoint := fmt.Sprintf("/playlists/%s/tracks", playlist.ID)
	if err := spotifyPost(accessToken, addEndpoint, bytes.NewReader(addBody), nil); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Playlist created but failed to add tracks: %v", err)})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Playlist created successfully",
		"data": gin.H{
			"playlist_id":  playlist.ID,
			"playlist_name": playlist.Name,
			"spotify_url":  playlist.ExternalURLs.Spotify,
			"track_count":  len(req.TrackURIs),
		},
	})
}
