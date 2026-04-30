// Package cache, uygulamanın Redis bağlantısını ve cache odaklı
// yardımcılarını içerir. Bağlantı tek bir yerde (main.go) açılır ve
// ihtiyacı olan modüllere (middleware, gelecekteki orchestrator
// önbelleği, vb.) *redis.Client olarak enjekte edilir.
//
// Modular monolith disiplinine uygun olarak Redis client'i da
// uygulama yaşam döngüsü boyunca paylaşılan tek bir instance'tır.
package cache

import (
	"context"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

// Varsayılan Redis konfigürasyonu — env değişkenleri tanımlı değilse
// kullanılır. Yerel geliştirmede `docker run redis` çıktısıyla uyumludur.
const (
	defaultAddr = "localhost:6379"
	defaultDB   = 0

	// pingTimeout, bağlantı doğrulama için kısa bir timeout. Redis
	// erişilemezse uygulama hızla fail-fast yapsın diye düşük tutuldu.
	pingTimeout = 5 * time.Second
)

// ConnectRedis, env değişkenlerini okuyarak Redis'e bağlanır ve
// PING ile bağlantıyı doğrular. Bağlanılamıyorsa uygulamayı durdurur
// (Postgres bağlantısı ile aynı fail-fast prensibi).
//
// Okunan env değişkenleri:
//
//	REDIS_ADDR      — host:port (varsayılan: localhost:6379)
//	REDIS_PASSWORD  — şifre (boş geçilebilir)
//	REDIS_DB        — DB indeksi (varsayılan: 0)
func ConnectRedis() *redis.Client {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = defaultAddr
	}

	password := os.Getenv("REDIS_PASSWORD")

	dbIdx := defaultDB
	if v := os.Getenv("REDIS_DB"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			dbIdx = n
		}
	}

	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       dbIdx,
	})

	// Bağlantıyı kısa bir context ile doğrula — uzun süre asılı kalmasın.
	ctx, cancel := context.WithTimeout(context.Background(), pingTimeout)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		log.Fatalf("❌ Failed to connect to Redis (%s): %v", addr, err)
	}

	log.Printf("✅ Redis connection established successfully (%s, db=%d)", addr, dbIdx)
	return client
}
