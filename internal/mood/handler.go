package mood

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// Handler, Mood modülünün HTTP katmanıdır.
// Yalnızca request/response dönüşümü ve hata kodu eşlemesi yapar;
// iş kuralları Service katmanına delege edilir.
type Handler struct {
	service *Service
}

// NewHandler, Service'i alarak yeni bir Mood handler üretir.
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// userIDFromContext, AuthMiddleware tarafından context'e konulan userID
// değerini güvenli şekilde çıkarır. Middleware doğru kuruluysa burada
// hata oluşmamalıdır; yine de defansif kontrol bırakıldı.
func userIDFromContext(c *gin.Context) (uint, bool) {
	raw, exists := c.Get("userID")
	if !exists {
		return 0, false
	}
	id, ok := raw.(uint)
	return id, ok
}

// Create handles POST /api/v1/moods — kullanıcının ham ruh hali metnini
// kabul eder ve "pending" durumunda bir Mood kaydı oluşturur.
//
// AI analizi bu noktada yapılmaz — orchestrator pipeline'ı (sonraki
// sprint) bu kaydı alıp Python servisine gönderecek ve sonuçları
// UpdateAnalysis ile yazacaktır. Bu sayede HTTP katmanı AI'ın
// yavaşlığından/başarısızlığından bağımsız kalır.
func (h *Handler) Create(c *gin.Context) {
	userID, ok := userIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Yetkilendirme bilgisi bulunamadı"})
		return
	}

	var req CreateMoodRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	m, err := h.service.CreateRawMood(userID, req.Text)
	if err != nil {
		if errors.Is(err, ErrEmptyText) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ruh hali kaydedilemedi"})
		return
	}

	// 202 Accepted — kayıt alındı, AI analizi arka planda tamamlanacak.
	c.JSON(http.StatusAccepted, gin.H{
		"message": "Ruh hali kaydedildi, analiz bekleniyor",
		"data":    m,
	})
}

// GetByID handles GET /api/v1/moods/:id — tek bir Mood kaydı döner.
// Kayıt başka bir kullanıcıya aitse 404 döner (varlığı sızdırmamak için).
func (h *Handler) GetByID(c *gin.Context) {
	userID, ok := userIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Yetkilendirme bilgisi bulunamadı"})
		return
	}

	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz ruh hali ID'si"})
		return
	}

	m, err := h.service.GetMoodByID(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ruh hali getirilemedi"})
		return
	}
	if m == nil || m.UserID != userID {
		c.JSON(http.StatusNotFound, gin.H{"error": "Ruh hali bulunamadı"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Ruh hali başarıyla getirildi",
		"data":    m,
	})
}

// List handles GET /api/v1/moods — giriş yapmış kullanıcının kendi
// ruh hali geçmişini sayfalı olarak döner.
//
// Query parametreleri:
//
//	limit  — sayfa başına kayıt (varsayılan 20, maksimum 100)
//	offset — atlanacak kayıt sayısı (varsayılan 0)
func (h *Handler) List(c *gin.Context) {
	userID, ok := userIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Yetkilendirme bilgisi bulunamadı"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	moods, err := h.service.GetUserMoods(userID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ruh hali geçmişi getirilemedi"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Ruh hali geçmişi başarıyla getirildi",
		"data":    moods,
		"count":   len(moods),
	})
}
