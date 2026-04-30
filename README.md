# 🎵 Aria — RAG Tabanlı Müzik Kürasyonu Sistemi

Modular Monolith mimarisinde Go (Golang) ile geliştirilmiş bir müzik kürasyonu backend API'si.

## 🏗️ Teknoloji Stack

| Teknoloji | Kullanım |
|-----------|----------|
| Go 1.26 | Ana programlama dili |
| Gin | HTTP web framework |
| GORM | ORM (Object-Relational Mapping) |
| PostgreSQL | Ana veritabanı |
| JWT | Token tabanlı kimlik doğrulama |
| Bcrypt | Şifre hashleme |
| OAuth 2.0 | Google & Spotify kimlik doğrulama |
| Redis (go-redis v9) | Cache & rate limiting |
| Python FastAPI (harici) | Sentiment analizi & RAG tabanlı öneri motoru |


## 📁 Proje Yapısı

```
aria-backend/
├── cmd/
│   └── api/
│       └── main.go              # Uygulama giriş noktası
├── internal/
│   ├── auth/
│   │   ├── dto.go               # Request/Response yapıları
│   │   ├── handler.go           # Signup, Login, Me handler'ları
│   │   ├── oauth_google.go      # Google OAuth 2.0 config + handler
│   │   ├── oauth_spotify.go     # Spotify OAuth 2.0 config + handler

│   │   └── routes.go            # Auth route tanımları
│   ├── middleware/
│   │   ├── auth.go              # JWT doğrulama middleware'i
│   │   └── ratelimit.go         # Redis tabanlı rate limit middleware'i
│   ├── mood/
│   │   ├── dto.go               # Request/Response & internal DTO'lar
│   │   ├── handler.go           # Mood HTTP handler'ları
│   │   ├── model.go             # GORM Mood modeli (sentiment alanları + jsonb)
│   │   ├── repository.go        # Mood veritabanı erişim katmanı
│   │   ├── routes.go            # Mood route tanımları
│   │   └── service.go           # Mood iş mantığı katmanı
│   ├── recommendation/
│   │   ├── dto.go               # Internal DTO'lar (orchestrator giriş yapısı)
│   │   ├── handler.go           # Recommendation HTTP handler'ları (sadece okuma)
│   │   ├── model.go             # Recommendation + RecommendedTrack modelleri
│   │   ├── repository.go        # Recommendation veritabanı erişim katmanı
│   │   ├── routes.go            # Recommendation route tanımları
│   │   └── service.go           # Recommendation iş mantığı katmanı
│   ├── pipeline/
│   │   ├── dto.go               # GenerateRequest / GenerateResponse
│   │   ├── handler.go           # /recommendations/generate handler'ı
│   │   ├── routes.go            # Orchestrator route tanımları
│   │   └── service.go           # AI çağrıları + DB persist orkestrasyonu
│   ├── music/
│   │   └── model.go             # Spotify entegrasyonu (placeholder)
│   ├── seeder/
│   │   └── seeder.go            # Örnek kullanıcı seeder'ı
│   └── user/
│       ├── handler.go           # User HTTP handler'ları
│       ├── model.go             # GORM User modeli
│       ├── repository.go        # Veritabanı erişim katmanı
│       ├── routes.go            # User route tanımları
│       └── service.go           # İş mantığı katmanı
├── pkg/
│   ├── aiclient/
│   │   ├── client.go            # Python AI servisi HTTP istemcisi
│   │   └── contract.go          # /analyze ve /recommend JSON sözleşmesi (spec)
│   ├── cache/
│   │   └── redis.go             # Redis bağlantısı (go-redis v9)
│   ├── database/
│   │   └── postgres.go          # PostgreSQL bağlantısı
│   └── utils/
│       ├── jwt.go               # JWT üretimi ve doğrulaması
│       └── password.go          # Bcrypt hash/kontrol
├── .env.example                 # Örnek ortam değişkenleri
├── .gitignore
├── go.mod
└── go.sum
```

## 🚀 Kurulum

### Gereksinimler

- Go 1.22+
- PostgreSQL (veya Docker)
- Redis 7+ (rate limiting için zorunlu — Docker ile çalıştırılabilir)
- (Opsiyonel) Python FastAPI AI servisi — `AI_SERVICE_URL` ile erişilebilir olmalı; servis ayağa kalkmadan `/recommendations/generate` 503 döner.

### Adımlar

**1. Repoyu klonla:**
```bash
git clone https://github.com/KULLANICI_ADIN/aria-backend.git
cd aria-backend
```

**2. Ortam değişkenlerini ayarla:**

```bash
cp .env.example .env
```

`.env` dosyasını düzenleyip kendi bilgilerini gir.

**3. PostgreSQL'i başlat (Docker ile):**

```bash
docker run -d \
  --name aria-postgres \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=postgres123 \
  -e POSTGRES_DB=music_curation \
  -p 5432:5432 \
  postgres:16-alpine
```

Veya manuel olarak veritabanını oluştur:

```sql
CREATE DATABASE music_curation;
```

**4. Redis'i başlat (Docker ile):**

```bash
docker run -d \
  --name aria-redis \
  -p 6379:6379 \
  redis:7-alpine
```

> Redis bağlantısı zorunludur — `pkg/cache.ConnectRedis` PING başarısız olursa uygulama fail-fast durur. Bu, rate limiter'ın güvenilir çalışması için bilinçli bir tercihtir.

**5. Bağımlılıkları indir:**
```bash
go mod tidy
```

**6. Sunucuyu başlat:**
```bash
go run cmd/api/main.go
```

Sunucu `http://localhost:8080` adresinde çalışmaya başlayacak.

## 📡 API Endpoint'leri

### Auth — Public


| Method | Endpoint | Açıklama |
|--------|----------|----------|
| `POST` | `/api/v1/auth/signup` | Yeni kullanıcı kaydı |
| `POST` | `/api/v1/auth/login` | Giriş yap, JWT al |
| `GET` | `/api/v1/auth/me` | Mevcut kullanıcıyı getir 🔒 |

### Auth — OAuth 2.0

| Method | Endpoint | Açıklama |
|--------|----------|----------|
| `GET` | `/api/v1/auth/google` | Google ile giriş — consent ekranına yönlendirir |
| `GET` | `/api/v1/auth/google/callback` | Google callback — JWT döner |
| `GET` | `/api/v1/auth/spotify` | Spotify ile giriş — authorize sayfasına yönlendirir |
| `GET` | `/api/v1/auth/spotify/callback` | Spotify callback — JWT döner |

### Users — Protected 🔒

| Method | Endpoint | Açıklama |
|--------|----------|----------|
| `GET` | `/api/v1/users` | Tüm kullanıcıları listele |
| `GET` | `/api/v1/users/:id` | ID ile kullanıcı getir |

### Mood — Protected 🔒

Kullanıcının ham ruh hali metinlerini ve bunların AI tarafından üretilmiş sentiment skorlarını saklar.

| Method | Endpoint | Açıklama |
|--------|----------|----------|
| `POST` | `/api/v1/moods` | Ham metin kaydeder, `pending` durumunda Mood döner (AI çalıştırmaz) |
| `GET` | `/api/v1/moods` | Giriş yapmış kullanıcının ruh hali geçmişi (sayfalı: `?limit=&offset=`) |
| `GET` | `/api/v1/moods/:id` | Tek bir Mood kaydını getirir |
| `GET` | `/api/v1/moods/:id/recommendations` | Bu Mood için üretilmiş tüm öneri kümeleri |

> **Not:** Asıl uçtan uca akış (sentiment + öneri) `/api/v1/recommendations/generate` üzerinden yapılır. `POST /moods` yalnızca AI'sız ham kayıt için kullanılır; üretimde nadiren çağrılması beklenir.

### Recommendation — Protected 🔒

| Method | Endpoint | Açıklama |
|--------|----------|----------|
| `POST` | `/api/v1/recommendations/generate` | Tüm pipeline'ı çalıştırır — sentiment analizi + RAG önerisi (rate limited) |
| `GET` | `/api/v1/recommendations` | Kullanıcının öneri geçmişi (sayfalı, parçalar dahil değil) |
| `GET` | `/api/v1/recommendations/:id` | Tek bir öneri kümesi (parçalarıyla, Position'a göre sıralı) |

> **Yazma yok (POST/PUT/DELETE):** Recommendation kayıtları **yalnızca** orchestrator (`pipeline.Service.CreateFromAI`) üzerinden oluşturulur. Bu, AI sonuçlarının dışarıdan elle manipüle edilmesini engelleyen bilinçli bir mimari karardır.

### Health

| Method | Endpoint | Açıklama |
|--------|----------|----------|
| `GET` | `/health` | API sağlık kontrolü |

## 📝 Kullanım Örnekleri

**Kayıt ol:**
```bash
curl -X POST http://localhost:8080/api/v1/auth/signup \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"123456"}'
```

**Giriş yap:**
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"123456"}'
```

**Korumalı endpoint'e erişim:**
```bash
curl http://localhost:8080/api/v1/users \
  -H "Authorization: Bearer BURAYA_JWT_TOKEN"
```

**Google ile giriş:**
Tarayıcıda aç: `http://localhost:8080/api/v1/auth/google`

**Spotify ile giriş:**
Tarayıcıda aç: `http://127.0.0.1:8080/api/v1/auth/spotify`

## 🔐 OAuth 2.0 Entegrasyonu

### Google OAuth 2.0

- Kullanıcı `/api/v1/auth/google` → Google consent ekranına yönlendirilir
- Giriş sonrası callback'e döner → Google'dan email ve google_id alınır
- DB'de kullanıcı varsa eşleştirilir, yoksa yeni oluşturulur → JWT döner

**Gerekli env değişkenleri:**
```
GOOGLE_CLIENT_ID=...
GOOGLE_CLIENT_SECRET=...
GOOGLE_REDIRECT_URL=http://localhost:8080/api/v1/auth/google/callback
```

### Spotify OAuth 2.0

- Kullanıcı `/api/v1/auth/spotify` → Spotify authorize sayfasına yönlendirilir
- Giriş sonrası callback'e döner → email, spotify_id, access_token, refresh_token alınır
- DB'de kullanıcı varsa eşleştirilir ve token'lar güncellenir, yoksa yeni oluşturulur → JWT döner
- Spotify `access_token` ve `refresh_token` DB'ye kaydedilir (Spotify API çağrıları için)

**Gerekli env değişkenleri:**
```
SPOTIFY_CLIENT_ID=...
SPOTIFY_CLIENT_SECRET=...
SPOTIFY_REDIRECT_URL=http://127.0.0.1:8080/api/v1/auth/spotify/callback
```

> **Not:** Spotify, `localhost` yerine `127.0.0.1` kullanılmasını gerektiriyor. Test ederken tarayıcıda da `127.0.0.1` kullanın.

**Spotify OAuth Scopes:**
- `user-read-email` — kullanıcı emaili
- `user-read-private` — kullanıcı profili
- `playlist-modify-public` — public playlist oluşturma
- `playlist-modify-private` — private playlist oluşturma

### OAuth Kullanıcı Eşleştirme Mantığı

1. Provider ID (google_id / spotify_id) ile DB'de arama yapılır
2. Bulunursa → mevcut kullanıcı döner
3. Bulunamazsa → email ile arama yapılır
4. Email varsa → provider ID mevcut hesaba bağlanır
5. Email de yoksa → yeni kullanıcı oluşturulur

## 🧠 Orchestrator Akışı — `POST /api/v1/recommendations/generate`

Bu endpoint, sistemin **uçtan uca tek girişi**dir: kullanıcının ham metnini alır, sentiment analizinden parça önerisine kadar tüm boru hattını (pipeline) tek bir senkron çağrıda çalıştırır.

**İstek:**

```bash
curl -X POST http://localhost:8080/api/v1/recommendations/generate \
  -H "Authorization: Bearer JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "text": "Bugün bahar gibiyim, biraz dansa ihtiyacım var",
    "limit": 20
  }'
```

`limit` opsiyoneldir (1–50 arası, varsayılan 20). `text` zorunlu, 1–2000 karakter.

**Akış (`internal/pipeline/service.go`):**

```
┌─────────────────────────────────────────────────────────────────┐
│ Frontend  →  POST /recommendations/generate                     │
└─────────┬───────────────────────────────────────────────────────┘
          │ AuthMiddleware (JWT)
          │ RateLimitMiddleware (Redis: 5/dk)
          ▼
┌─────────────────────────────────────────────────────────────────┐
│ pipeline.Service.GeneratePlaylist                               │
│                                                                 │
│ (a) mood.CreateRawMood        → DB'ye 'pending' kayıt           │
│ (b) aiclient.AnalyzeMood      → POST /analyze (Python)          │
│ (c) mood.UpdateAnalysis       → DB güncellenir, status=analyzed │
│ (d) aiclient.GetRecommendations → POST /recommend (Python)      │
│ (e) recommendation.CreateFromAI → tek transaction'da rec+tracks │
│ (f) GenerateResponse{Mood, Recommendation} döner                │
└─────────────────────────────────────────────────────────────────┘
```

Tüm modüller arası iletişim **direkt Go fonksiyon çağrısıyla** yapılır (modular monolith); ağ üzerinden iç çağrı yoktur. Sadece (b) ve (d) dış HTTP istekleridir (Python servisi).

**Hata davranışı (kısmi başarı toleranslı):**

| Adım | Hata olursa | Mood durumu |
|------|-------------|-------------|
| (a) | DB hatası, hiç kayıt kalmaz | — |
| (b) | AI servisi yanıt vermedi | `failed` olarak işaretlenir |
| (c) | DB update başarısız | `pending` kalır (idempotent retry mümkün) |
| (d) | AI servisi yanıt vermedi | `analyzed` kalır, recommendation yazılmaz |
| (e) | DB transaction başarısız | `analyzed` kalır |

**HTTP yanıt kodları:**

| Kod | Anlam |
|-----|-------|
| `200 OK` | Başarılı; `{ mood, recommendation: { tracks: [...] } }` döner |
| `400 Bad Request` | Boş/aşırı uzun metin, geçersiz `limit` |
| `401 Unauthorized` | JWT yok veya geçersiz |
| `429 Too Many Requests` | Rate limit aşıldı (`Retry-After` header'ı ile) |
| `502 Bad Gateway` | AI servisi 4xx/5xx döndürdü (`code: ai_bad_request` / `ai_internal`) |
| `503 Service Unavailable` | AI servisine ulaşılamadı (`code: ai_unavailable`) |
| `500 Internal Server Error` | Veritabanı veya beklenmeyen hata |

## 🚦 Rate Limiting

`POST /api/v1/recommendations/generate` endpoint'i, AI servisinin maliyetli olması nedeniyle Redis tabanlı sabit pencere (fixed window) rate limit ile korunur.

| Kapsam | Limit | Pencere |
|--------|-------|---------|
| Kullanıcı başına `/recommendations/generate` | **5 istek** | **1 dakika** |

Diğer endpoint'lerde (mood, recommendation listeleme, auth, vb.) rate limit **yoktur**.

**Anahtar şeması (Redis):**

```
ratelimit:generate:<userID>      → INCR sayacı (TTL = 60s)
```

**İstemci tarafı yardımcı header'lar:**

| Header | Anlam |
|--------|-------|
| `X-RateLimit-Limit` | Pencere başına izin verilen toplam istek (5) |
| `X-RateLimit-Remaining` | Bu pencerede kalan istek hakkı |
| `Retry-After` | (sadece 429'da) Pencerenin sıfırlanmasına kalan saniye |

**Fail-open politikası:** Redis erişilemez durumdaysa rate limit middleware isteği **bloklamaz**, sunucu loguna uyarı yazar ve geçirir. Redis blip'i tüm API'yi çökertmesin diye bilinçli bir tercihtir; auth zinciri her zaman zorlu olarak çalışmaya devam eder.

## 🔌 Python AI Servisi Sözleşmesi

Backend'in beklediği JSON şemaları `pkg/aiclient/contract.go` dosyasında **tek doğru kaynak** (single source of truth) olarak tanımlıdır. Python ekibi Pydantic modellerini bu struct'larla birebir uyumlu yazmalıdır.

| Endpoint | Amaç | Request | Response |
|----------|------|---------|----------|
| `POST {AI_SERVICE_URL}/analyze` | Sentiment + duygu analizi | `AnalyzeRequest` | `AnalyzeResponse` |
| `POST {AI_SERVICE_URL}/recommend` | RAG tabanlı parça önerisi | `RecommendRequest` | `RecommendResponse` |

Detaylı alan açıklamaları, status kodları ve örnek payload'lar için `pkg/aiclient/contract.go` dosyasına bakınız.

## 🗃️ Veritabanı Şeması

### Users Tablosu

| Alan | Tip | Açıklama |
|------|-----|----------|
| id | uint (PK) | Otomatik artan primary key |
| email | string (Unique) | Kullanıcı e-postası |
| password | string | Bcrypt ile hashlenmiş şifre (OAuth için opsiyonel) |
| google_id | string? | Google OAuth ID (nullable) |
| spotify_id | string? | Spotify OAuth ID (nullable) |
| spotify_access_token | string? | Spotify API erişim token'ı (nullable) |
| spotify_refresh_token | string? | Spotify token yenileme (nullable) |
| created_at | timestamp | Oluşturulma tarihi |
| updated_at | timestamp | Güncellenme tarihi |

### Moods Tablosu

| Alan | Tip | Açıklama |
|------|-----|----------|
| id | uint (PK) | Otomatik artan primary key |
| user_id | uint (FK, indexed) | Sahip kullanıcı |
| raw_text | text | Kullanıcının girdiği ham metin |
| sentiment_label | string | AI etiketi (pozitif/negatif/karışık vb.) |
| dominant_emotion | string | En yüksek skorlu duygu |
| valence | float | [-1, +1] — pozitif/negatif duygulanım |
| arousal | float | [-1, +1] — sakin/heyecanlı |
| energy | float | [0, 1] — Spotify energy benzeri |
| emotion_scores | jsonb | Çok-etiketli duygu skor haritası |
| language | string | ISO 639-1 dil kodu |
| ai_model_version | string | Analizi üreten model sürümü |
| processing_ms | int | AI çağrısı süresi (ms) |
| status | string (indexed) | `pending` / `analyzed` / `failed` |
| created_at, updated_at | timestamp | — |

### Recommendations Tablosu

| Alan | Tip | Açıklama |
|------|-----|----------|
| id | uint (PK) | Otomatik artan primary key |
| user_id | uint (FK, indexed) | Sahip kullanıcı |
| mood_id | uint (FK, indexed) | İlişkili Mood kaydı |
| ai_model_version | string | Öneriyi üreten RAG modelinin sürümü |
| rag_context | text | LLM prompt'u veya retrieve edilen bağlam (debug/explainability) |
| processing_ms | int | AI çağrısı süresi (ms) |
| status | string (indexed) | `pending` / `ready` / `failed` |
| created_at, updated_at | timestamp | — |

### RecommendedTracks Tablosu

| Alan | Tip | Açıklama |
|------|-----|----------|
| id | uint (PK) | Otomatik artan primary key |
| recommendation_id | uint (FK, indexed) | Parent Recommendation; CASCADE delete |
| spotify_track_id | string (indexed, nullable) | Spotify katalog ID'si — sonradan da çözümlenebilir |
| title, artist, album | string | Parça meta verisi |
| preview_url, external_url | string | Spotify preview ve open.spotify.com linkleri |
| duration_ms | int | Parça süresi (ms) |
| position | int | Küme içindeki sıra (0'dan başlar) |
| relevance_score | float | [0, 1] — AI'nın güven skoru |
| reason | text | "Neden bu şarkı?" açıklaması |
| created_at | timestamp | — |

> Composite unique index: `(recommendation_id, position)` — aynı küme içinde Position çakışmasını engeller.

## 🌱 Seeder

Uygulama ilk başlatıldığında otomatik olarak 10 örnek kullanıcı oluşturur:

| Email | Şifre |
|-------|-------|
| alice@example.com | password123 |
| bob@example.com | password123 |
| charlie@example.com | password123 |
| ... | ... |

Seeder idempotent'tır — tekrar çalıştırıldığında mevcut kullanıcıları tekrar eklemez.

## 🔜 Yol Haritası

- [x] Proje iskeleti & Auth (signup, login, me)
- [x] Google OAuth 2.0 entegrasyonu
- [x] Spotify OAuth 2.0 (Access & Refresh Token yönetimi)
- [x] Mood + Recommendation modülleri (GORM + jsonb)
- [x] AI/RAG client + Python servisi sözleşmesi (`pkg/aiclient`)
- [x] Orchestrator pipeline (`POST /recommendations/generate`)
- [x] Redis bağlantısı + per-user rate limiting (5/dk)
- [ ] Spotify API ile gerçek playlist oluşturma (parça çözümleme)
- [ ] Öneri sonuçlarının Redis ile cache'lenmesi
- [ ] Asenkron pipeline (job queue + status polling)

## 👥 Ekip

| Görev | Durum |
|-------|-------|
| Proje iskeleti & Auth | ✅ Tamamlandı |
| Google/Spotify OAuth | ✅ Tamamlandı |
| AI Integration & Data Layer (mood, recommendation, pipeline, aiclient, cache) | ✅ Tamamlandı |
| Spotify API entegrasyonu (playlist push) | 🔜 Geliyor |
| Python AI/RAG servisi | 🔜 Geliyor |

## 📄 Lisans

Bu proje özel kullanım içindir.