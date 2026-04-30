package recommendation

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// Handler, Recommendation modülünün HTTP katmanıdır.
// Yalnızca okuma (GET) endpoint'leri sunar — yazma işlemleri orchestrator
// pipeline'ı üzerinden Service.CreateFromAI ile yapılır. Bu sayede dış
// dünyaya AI sonuçlarını "elle uydurarak" ekleme yolu kapatılmış olur.
type Handler struct {
	service *Service
}

// NewHandler, Service'i alarak yeni bir Recommendation handler üretir.
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

// GetByID handles GET /api/v1/recommendations/:id — tek bir öneri
// kümesini parçalarıyla birlikte döner. Başka kullanıcının kaydıysa
// 404 verir (varlığı sızdırmamak için).
func (h *Handler) GetByID(c *gin.Context) {
	userID, ok := userIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Yetkilendirme bilgisi bulunamadı"})
		return
	}

	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz öneri ID'si"})
		return
	}

	rec, err := h.service.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Öneri getirilemedi"})
		return
	}
	if rec == nil || rec.UserID != userID {
		c.JSON(http.StatusNotFound, gin.H{"error": "Öneri bulunamadı"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Öneri başarıyla getirildi",
		"data":    rec,
	})
}

// List handles GET /api/v1/recommendations — kullanıcının kendi öneri
// geçmişini sayfalı olarak döner. Performans için Track listesi dahil
// edilmez; detay için /recommendations/:id kullanılır.
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

	recs, err := h.service.GetUserRecommendations(userID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Öneri geçmişi getirilemedi"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Öneri geçmişi başarıyla getirildi",
		"data":    recs,
		"count":   len(recs),
	})
}

// GetByMoodID handles GET /api/v1/moods/:id/recommendations — belirli
// bir Mood için üretilmiş tüm öneri kümelerini döner.
//
// Yetki kontrolü: önce öneri sahibinin kullanıcı olduğunu doğrular.
// Mood'a doğrudan join yapmıyoruz; bunun yerine dönen önerilerin
// UserID'sini kontrol ediyoruz. Mood başka bir kullanıcıya aitse boş
// liste dönecektir, bu da 404'e benzer bir davranış sağlar.
func (h *Handler) GetByMoodID(c *gin.Context) {
	userID, ok := userIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Yetkilendirme bilgisi bulunamadı"})
		return
	}

	moodIDParam := c.Param("id")
	moodID, err := strconv.ParseUint(moodIDParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz ruh hali ID'si"})
		return
	}

	all, err := h.service.GetByMoodID(uint(moodID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Öneriler getirilemedi"})
		return
	}

	// Yalnızca giriş yapmış kullanıcıya ait öneriler döndürülür.
	owned := make([]Recommendation, 0, len(all))
	for _, r := range all {
		if r.UserID == userID {
			owned = append(owned, r)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Ruh hali için öneriler getirildi",
		"data":    owned,
		"count":   len(owned),
	})
}
