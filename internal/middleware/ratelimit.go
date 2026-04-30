package middleware

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// RateLimitMiddleware, Redis tabanlı sabit pencereli (fixed window)
// rate limiting middleware'i üretir. Üst katmandaki AuthMiddleware
// userID'yi context'e koymuş olmalıdır — bu middleware ondan SONRA
// zincirlenmelidir.
//
// Algoritma (sabit pencere):
//
//	key   = "ratelimit:<scope>:<userID>"
//	count = INCR(key)
//	if count == 1 → EXPIRE(key, window)
//	if count > limit → 429 Too Many Requests
//
// Sabit pencere, sliding window'dan daha basittir ve hafif şekilde
// "sınırı kötüye kullanma" alanı bırakır (pencere bitiminde aniden
// sıfırlanır), ancak Redis tarafında tek atomik komutla çalışır ve
// 5/dakika gibi düşük limitler için fazlasıyla yeterlidir.
//
// Parametreler:
//
//	rdb    — paylaşılan Redis client (main.go'da yaratılır)
//	scope  — limit anahtarındaki ayraç (örn. "generate", "search")
//	limit  — pencere başına izin verilen istek sayısı
//	window — pencere uzunluğu (örn. time.Minute)
//
// Hata politikası:
//
//	Redis erişilemezse fail-OPEN davranılır (uyarı log'lanır, istek
//	geçirilir). Sebep: Redis blip'i uygulamanın tamamını çökertmesin.
//	Auth her zaman aktif kalmaya devam eder; rate limit sadece bir
//	abuse koruması katmanıdır, single source of truth değildir.
func RateLimitMiddleware(rdb *redis.Client, scope string, limit int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDRaw, exists := c.Get("userID")
		if !exists {
			// AuthMiddleware bağlanmamışsa veya başarısızsa.
			// Yine de defansif kontrol; konfig hatasını yakalar.
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Yetkilendirme bilgisi bulunamadı"})
			c.Abort()
			return
		}
		userID, ok := userIDRaw.(uint)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Yetkilendirme bilgisi geçersiz"})
			c.Abort()
			return
		}

		key := fmt.Sprintf("ratelimit:%s:%d", scope, userID)

		// Redis çağrılarını request context'inden bağımsız, kısa bir
		// timeout ile yap — request iptal edilse bile sayaç yine de
		// güncellenmiş olur (state tutarlılığı için).
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		count, err := rdb.Incr(ctx, key).Result()
		if err != nil {
			// Fail-open: log'la ve isteği geçir. Redis kapalıyken
			// kullanıcıyı dışarıda bırakmak istemiyoruz.
			log.Printf("⚠️  rate limit kontrolü başarısız (fail-open): scope=%s userID=%d err=%v", scope, userID, err)
			c.Next()
			return
		}

		// Sayaç yeni oluşturuldu — pencere TTL'i ata.
		// EXPIRE çağrısı başarısız olsa bile (race condition), sayaç
		// max ~Redis'in bekleme süresi kadar kalır; bir sonraki INCR
		// tekrar EXPIRE'ı tetikler. Yine de hatayı log'luyoruz.
		if count == 1 {
			if err := rdb.Expire(ctx, key, window).Err(); err != nil {
				log.Printf("⚠️  rate limit TTL ayarlanamadı: key=%s err=%v", key, err)
			}
		}

		if count > int64(limit) {
			// Pencerenin ne kadar kaldığını öğren — Retry-After header'ı
			// için. TTL alınamazsa pencerenin tamamı kadar varsay.
			ttl, ttlErr := rdb.TTL(ctx, key).Result()
			if ttlErr != nil || ttl < 0 {
				ttl = window
			}
			retryAfter := int(ttl.Seconds())
			if retryAfter < 1 {
				retryAfter = 1
			}

			c.Header("Retry-After", fmt.Sprintf("%d", retryAfter))
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":               "İstek limiti aşıldı, lütfen biraz sonra tekrar deneyin",
				"limit":               limit,
				"window_seconds":      int(window.Seconds()),
				"retry_after_seconds": retryAfter,
			})
			c.Abort()
			return
		}

		// Limit dahilinde — isteği bir sonraki handler'a geçir.
		// Kalan kontör sayısını response header'ında bilgi olarak ver.
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", limit))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", int64(limit)-count))

		c.Next()
	}
}
