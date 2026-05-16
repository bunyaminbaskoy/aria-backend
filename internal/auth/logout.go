package auth

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"music-curation/pkg/utils"
)

// blacklistClient, token kara listesi için kullanılan Redis istemcisi.
// InitBlacklist çağrılmadıysa nil kalır; bu durumda in-process fallback
// devre dışı bırakılır ve her token geçerli sayılır (güvenli-fail-open).
var blacklistClient *redis.Client

// InitBlacklist, Redis istemcisini auth paketine enjekte eder.
// main.go içinde uygulama başlangıcında tek seferlik çağrılır.
func InitBlacklist(rdb *redis.Client) {
	blacklistClient = rdb
	log.Println("✅ Auth token blacklist Redis'e bağlandı")
}

// blacklistKey, token için deterministik bir Redis anahtarı üretir.
func blacklistKey(token string) string {
	return fmt.Sprintf("auth:blacklist:%s", token)
}

// BlacklistToken — Token'ı Redis'te kara listeye ekle.
// TTL, token'ın gerçek son kullanma süresine göre ayarlanır;
// bu sayede Redis süresi dolan girişleri otomatik siler.
func BlacklistToken(token string, expiresAt time.Time) {
	if blacklistClient == nil {
		log.Println("⚠️  Blacklist Redis istemcisi başlatılmadı, token iptal edilemedi")
		return
	}

	ttl := time.Until(expiresAt)
	if ttl <= 0 {
		// Token zaten süresi dolmuş — kara listeye eklemeye gerek yok.
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	key := blacklistKey(token)
	if err := blacklistClient.Set(ctx, key, "1", ttl).Err(); err != nil {
		log.Printf("⚠️  Token kara listeye eklenemedi (%s): %v", key, err)
	}
}

// IsTokenBlacklisted — Token Redis kara listesinde mi?
func IsTokenBlacklisted(token string) bool {
	if blacklistClient == nil {
		return false
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	exists, err := blacklistClient.Exists(ctx, blacklistKey(token)).Result()
	if err != nil {
		log.Printf("⚠️  Blacklist sorgusu başarısız: %v", err)
		// Hata durumunda güvenli-fail-open: token geçerli say.
		return false
	}
	return exists > 0
}

// CleanupBlacklist — Redis otomatik TTL silme kullandığından
// artık manuel temizlik gerekmiyor. Eski çağrı noktalarıyla
// geriye dönük uyumluluk için boş bırakıldı.
//
// Deprecated: Redis TTL temizliği otomatik yapar.
func CleanupBlacklist() {}

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
