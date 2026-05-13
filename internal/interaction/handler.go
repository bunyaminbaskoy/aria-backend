package interaction

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// Handler, TrackInteraction modülünün HTTP katmanıdır.
type Handler struct {
	service *Service
}

// NewHandler, Service'i alarak yeni bir interaction handler üretir.
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// userIDFromContext, AuthMiddleware tarafından context'e konulan
// userID'yi güvenli şekilde çıkarır.
func userIDFromContext(c *gin.Context) (uint, bool) {
	raw, exists := c.Get("userID")
	if !exists {
		return 0, false
	}
	id, ok := raw.(uint)
	return id, ok
}

// Create handles POST /api/v1/interactions — kullanıcının bir parçayla
// etkileşimini (like/dislike) kaydeder.
func (h *Handler) Create(c *gin.Context) {
	userID, ok := userIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Yetkilendirme bilgisi bulunamadı"})
		return
	}

	var req CreateInteractionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz istek: " + err.Error()})
		return
	}

	interaction, err := h.service.Upsert(userID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Etkileşim kaydedilemedi"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Etkileşim başarıyla kaydedildi",
		"data":    interaction,
	})
}

// List handles GET /api/v1/interactions — kullanıcının etkileşim
// geçmişini sayfalı olarak döner.
func (h *Handler) List(c *gin.Context) {
	userID, ok := userIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Yetkilendirme bilgisi bulunamadı"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	interactions, err := h.service.GetUserInteractions(userID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Etkileşimler getirilemedi"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Etkileşimler başarıyla getirildi",
		"data":    interactions,
		"count":   len(interactions),
	})
}

// Delete handles DELETE /api/v1/interactions/:spotify_track_id —
// kullanıcının belirli bir parçayla etkileşimini siler.
func (h *Handler) Delete(c *gin.Context) {
	userID, ok := userIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Yetkilendirme bilgisi bulunamadı"})
		return
	}

	spotifyTrackID := c.Param("spotify_track_id")
	if spotifyTrackID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Spotify track ID gerekli"})
		return
	}

	if err := h.service.DeleteInteraction(userID, spotifyTrackID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Etkileşim silinemedi"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Etkileşim başarıyla silindi",
	})
}
