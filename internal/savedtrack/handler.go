package savedtrack

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func userID(c *gin.Context) (uint, bool) {
	raw, ok := c.Get("userID")
	if !ok {
		return 0, false
	}
	id, ok := raw.(uint)
	return id, ok
}

func (h *Handler) Save(c *gin.Context) {
	uid, ok := userID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Yetkilendirme gerekli"})
		return
	}
	var req SaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	track, err := h.service.Save(uid, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Kaydetme başarısız"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": track})
}

func (h *Handler) List(c *gin.Context) {
	uid, ok := userID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Yetkilendirme gerekli"})
		return
	}
	moodKey := c.Query("mood")
	var tracks []SavedTrack
	var err error
	if moodKey != "" {
		tracks, err = h.service.GetUserSavedByMood(uid, moodKey)
	} else {
		tracks, err = h.service.GetUserSaved(uid)
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Kaydedilenler yüklenemedi"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": tracks})
}

func (h *Handler) Delete(c *gin.Context) {
	uid, ok := userID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Yetkilendirme gerekli"})
		return
	}
	spotifyID := c.Param("spotify_track_id")
	if spotifyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Spotify track ID gerekli"})
		return
	}
	if err := h.service.Unsave(uid, spotifyID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Silme başarısız"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Kayıt silindi"})
}
