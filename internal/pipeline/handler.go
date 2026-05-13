package pipeline

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// Handler, pipeline modülünün HTTP katmanıdır. Yalnızca request
// doğrulama, context'ten userID çekme ve hata kategorilerini HTTP
// status koduna eşleme yapar.
type Handler struct {
	service *Service
}

// NewHandler, Service'i alarak yeni bir orchestrator handler üretir.
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// userIDFromContext, AuthMiddleware tarafından context'e konulan
// userID değerini güvenli şekilde çıkarır.
func userIDFromContext(c *gin.Context) (uint, bool) {
	raw, exists := c.Get("userID")
	if !exists {
		return 0, false
	}
	id, ok := raw.(uint)
	return id, ok
}

// Generate handles POST /api/v1/recommendations/generate.
//
// Tüm sentiment analizi → öneri üretimi pipeline'ını çalıştıran tek
// endpoint'tir. AuthMiddleware ve RateLimitMiddleware bu handler'a
// gelmeden ÖNCE çalışır (route tarafında zincirlenir).
//
// Yanıtlar:
//
//	200 OK                    — başarılı; mood + recommendation döner
//	400 Bad Request           — boş/aşırı uzun metin, geçersiz limit
//	401 Unauthorized          — auth eksik (genellikle middleware yakalar)
//	429 Too Many Requests     — rate limit aşıldı (middleware yakalar)
//	502 Bad Gateway           — AI servisi 4xx döndürdü
//	503 Service Unavailable   — AI servisine ulaşılamadı
//	500 Internal Server Error — DB hatası veya beklenmeyen durum
func (h *Handler) Generate(c *gin.Context) {
	userID, ok := userIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Yetkilendirme bilgisi bulunamadı"})
		return
	}

	var req GenerateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if strings.TrimSpace(req.Text) == "" && strings.TrimSpace(req.MoodKey) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "text veya mood_key alanlarından en az biri gerekli"})
		return
	}

	result, err := h.service.GeneratePlaylist(c.Request.Context(), userID, req.Text, req.MoodKey, req.Limit)
	if err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Çalma listesi başarıyla oluşturuldu",
		"data":    result,
	})
}

// writeError, pipeline-level sentinel hatalarını uygun HTTP status
// koduna ve kullanıcı dostu Türkçe mesaja çevirir. Gerçek hata zinciri
// (örn. AI tarafının döndüğü ham mesaj) sunucu loguna yazılır;
// kullanıcıya iletilen mesaj kasıtlı olarak sade tutulur.
func writeError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrEmptyText):
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ruh hali metni boş olamaz"})

	case errors.Is(err, ErrAIUnavailable):
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error":   "AI servisi şu anda kullanılamıyor, lütfen birazdan tekrar deneyin",
			"code":    "ai_unavailable",
			"details": err.Error(),
		})

	case errors.Is(err, ErrAIBadRequest):
		// 502 Bad Gateway — backend ile AI arasındaki sözleşmede sorun var.
		c.JSON(http.StatusBadGateway, gin.H{
			"error":   "AI servisi isteği reddetti",
			"code":    "ai_bad_request",
			"details": err.Error(),
		})

	case errors.Is(err, ErrAIInternal):
		c.JSON(http.StatusBadGateway, gin.H{
			"error":   "AI servisinde dahili hata oluştu",
			"code":    "ai_internal",
			"details": err.Error(),
		})

	case errors.Is(err, ErrPersistence):
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Sonuç kaydedilemedi, lütfen tekrar deneyin",
			"code":  "persistence_error",
		})

	default:
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Beklenmeyen bir hata oluştu",
			"code":  "internal_error",
		})
	}
}
