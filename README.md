# 🎵 Aria — RAG Tabanlı Müzik Kürasyonu Sistemi

Modular Monolith mimarisinde Go (Golang) ile geliştirilmiş bir müzik kürasyonu backend API'si. Kullanıcının doğal dilde girdiği ruh hali metnini Python FastAPI (RAG/LLM) servisine gönderir, duygu analizi ve parça önerilerini yönetir, Spotify'a iletmeden önce tüm veri akışını düzenler.

---

## 🏗️ Teknoloji Stack

| Teknoloji | Kullanım |
|-----------|----------|
| Go 1.26 | Ana programlama dili |
| Gin | HTTP web framework |
| GORM | ORM (Object-Relational Mapping) |
| PostgreSQL | Ana veritabanı |
| JWT (golang-jwt/v5) | Token tabanlı kimlik doğrulama |
| Bcrypt | Şifre hashleme |
| OAuth 2.0 | Google & Spotify kimlik doğrulama |
| Redis (go-redis v9) | Cache & per-user rate limiting |
| Python FastAPI (harici) | Sentiment analizi & RAG tabanlı öneri motoru |

---

## 📁 Proje Yapısı

```
aria-backend/
├── cmd/
│   └── api/
│       └── main.go                  # Tek giriş noktası — tüm DI burada
├── internal/
│   ├── auth/
│   │   ├── dto.go                   # SignupRequest, LoginRequest, AuthResponse
│   │   ├── handler.go               # Signup, Login, Me
│   │   ├── logout.go                # Logout + token kara listesi (in-memory)
│   │   ├── oauth_google.go          # Google OAuth 2.0
│   │   ├── oauth_spotify.go         # Spotify OAuth 2.0 (token saklama dahil)
│   │   ├── refresh.go               # Access token yenileme
│   │   ├── routes.go                # Auth route tanımları
│   │   └── spotify_token.go         # SpotifyTokenManager (otomatik token yenileme)
│   ├── middleware/
│   │   ├── auth.go                  # JWT doğrulama middleware'i
│   │   ├── error.go                 # Global hata yakalama middleware'i
│   │   └── ratelimit.go             # Redis tabanlı fixed-window rate limiter
│   ├── mood/
│   │   ├── dto.go                   # CreateMoodRequest, AnalysisUpdate
│   │   ├── handler.go               # Mood HTTP handler'ları
│   │   ├── model.go                 # GORM Mood modeli (sentiment + jsonb)
│   │   ├── repository.go            # Mood DB erişim katmanı
│   │   ├── routes.go                # Mood route tanımları
│   │   └── service.go               # Mood iş mantığı
│   ├── recommendation/
│   │   ├── dto.go                   # CreateRecommendationInput, TrackInput
│   │   ├── handler.go               # Recommendation HTTP handler'ları (sadece okuma)
│   │   ├── model.go                 # Recommendation + RecommendedTrack modelleri
│   │   ├── repository.go            # Recommendation DB erişim katmanı
│   │   ├── routes.go                # Recommendation route tanımları
│   │   └── service.go               # Recommendation iş mantığı
│   ├── pipeline/
│   │   ├── dto.go                   # GenerateRequest / GenerateResponse
│   │   ├── handler.go               # /recommendations/generate handler'ı
│   │   ├── routes.go                # Orchestrator route tanımları
│   │   └── service.go               # 6-adımlı AI orchestration akışı
│   ├── spotify/
│   │   ├── client.go                # Spotify API HTTP yardımcıları (GET/POST)
│   │   ├── handler.go               # Spotify Handler yapısı
│   │   ├── history.go               # Son dinlenilen şarkılar
│   │   ├── playlist.go              # Playlist oluşturma & parça ekleme
│   │   ├── routes.go                # Spotify route tanımları
│   │   └── top.go                   # En çok dinlenen şarkılar & sanatçılar
│   ├── music/
│   │   └── model.go                 # Placeholder (kullanılmıyor)
│   ├── seeder/
│   │   └── seeder.go                # Örnek kullanıcı seeder'ı (idempotent)
│   └── user/
│       ├── handler.go               # User HTTP handler'ları
│       ├── model.go                 # GORM User modeli (local + OAuth alanları)
│       ├── repository.go            # User DB erişim katmanı
│       ├── routes.go                # User route tanımları
│       └── service.go               # User iş mantığı
├── pkg/
│   ├── aiclient/
│   │   ├── client.go                # Python AI servisi HTTP istemcisi (15s timeout)
│   │   └── contract.go              # /analyze ve /recommend JSON sözleşmesi (spec)
│   ├── cache/
│   │   └── redis.go                 # Redis bağlantısı (go-redis v9)
│   ├── database/
│   │   └── postgres.go              # PostgreSQL bağlantısı (GORM)
│   └── utils/
│       ├── jwt.go                   # JWT üretimi (access + refresh çifti) ve doğrulama
│       └── password.go              # Bcrypt hash/kontrol
├── Test/
│   ├── auth_jwt_test.go             # JWT, bcrypt, blacklist testleri
│   ├── mood_recommendation_test.go  # Model, DTO, sentinel hata testleri
│   ├── pipeline_service_test.go     # Mock AI client, hata eşleme testleri
│   └── spotify_module_test.go       # Spotify veri yapısı testleri
├── .env.example                     # Örnek ortam değişkenleri
├── .gitignore
├── go.mod
└── go.sum
```

---

## 🚀 Kurulum

### Gereksinimler

- Go 1.22+
- PostgreSQL 14+ (veya Docker)
- Redis 7+ (veya Docker)
- Python FastAPI AI servisi — `AI_SERVICE_URL` ile erişilebilir olmalı; servis ayağa kalkmadan `/recommendations/generate` 503 döner.

### Adımlar

**1. Repoyu klonla:**
```bash
git clone https://github.com/KULLANICI_ADIN/aria-backend.git
cd aria-backend
```

**2. Ortam değişkenlerini ayarla:**
```bash
cp .env.example .env
# .env dosyasını kendi bilgilerinle düzenle
```

**3. PostgreSQL'i başlat (Docker):**
```bash
docker run -d \
  --name aria-postgres \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=postgres123 \
  -e POSTGRES_DB=music_curation \
  -p 5432:5432 \
  postgres:16-alpine
```

**4. Redis'i başlat (Docker):**
```bash
docker run -d \
  --name aria-redis \
  -p 6379:6379 \
  redis:7-alpine
```

> Redis bağlantısı zorunludur. `ConnectRedis` PING başarısız olursa uygulama fail-fast durur.

**5. Bağımlılıkları indir:**
```bash
go mod tidy
```

**6. Sunucuyu başlat:**
```bash
go run cmd/api/main.go
```

Sunucu `http://localhost:8080` adresinde çalışmaya başlar ve AutoMigrate + Seeder otomatik çalışır.

---

## 📡 API Endpoint'leri

### Auth — Public

| Method | Endpoint | Açıklama |
|--------|----------|----------|
| `POST` | `/api/v1/auth/signup` | Yeni kullanıcı kaydı |
| `POST` | `/api/v1/auth/login` | Giriş yap, JWT çifti al |
| `POST` | `/api/v1/auth/refresh` | Refresh token ile yeni access token al |
| `POST` | `/api/v1/auth/logout` | Refresh token'ı kara listeye al |

### Auth — OAuth 2.0

| Method | Endpoint | Açıklama |
|--------|----------|----------|
| `GET` | `/api/v1/auth/me` | Mevcut kullanıcıyı getir 🔒 |
| `GET` | `/api/v1/auth/google` | Google consent ekranına yönlendir |
| `GET` | `/api/v1/auth/google/callback` | Google callback — JWT döner |
| `GET` | `/api/v1/auth/spotify` | Spotify authorize sayfasına yönlendir |
| `GET` | `/api/v1/auth/spotify/callback` | Spotify callback — JWT + token saklama |

### Users — Protected 🔒

| Method | Endpoint | Açıklama |
|--------|----------|----------|
| `GET` | `/api/v1/users` | Tüm kullanıcıları listele |
| `GET` | `/api/v1/users/:id` | ID ile kullanıcı getir |

### Mood — Protected 🔒

| Method | Endpoint | Açıklama |
|--------|----------|----------|
| `POST` | `/api/v1/moods` | Ham metin kaydeder, `pending` Mood döner (AI çalıştırmaz) |
| `GET` | `/api/v1/moods` | Kullanıcının ruh hali geçmişi (`?limit=&offset=`) |
| `GET` | `/api/v1/moods/:id` | Tek bir Mood kaydı |
| `GET` | `/api/v1/moods/:id/recommendations` | Bu Mood için üretilmiş öneriler |

> **Not:** Asıl uçtan uca akış `/api/v1/recommendations/generate` üzerinden yapılır. `POST /moods` AI'sız ham kayıt içindir.

### Recommendation — Protected 🔒

| Method | Endpoint | Açıklama |
|--------|----------|----------|
| `POST` | `/api/v1/recommendations/generate` | Pipeline: sentiment + RAG önerisi (rate limited: 5/dk) |
| `GET` | `/api/v1/recommendations` | Kullanıcının öneri geçmişi (sayfalı, parçasız) |
| `GET` | `/api/v1/recommendations/:id` | Tek öneri kümesi (parçalarıyla, Position sıralı) |

> Recommendation kayıtları **yalnızca** orchestrator üzerinden oluşturulur. Dışarıdan manuel POST yolu kapalıdır.

### Spotify — Protected 🔒

| Method | Endpoint | Açıklama |
|--------|----------|----------|
| `GET` | `/api/v1/spotify/history` | Son dinlenen şarkılar (`?limit=20`) |
| `GET` | `/api/v1/spotify/top/tracks` | En çok dinlenen şarkılar (`?time_range=&limit=`) |
| `GET` | `/api/v1/spotify/top/artists` | En çok dinlenen sanatçılar (`?time_range=&limit=`) |
| `POST` | `/api/v1/spotify/playlist` | Spotify'da playlist oluştur ve parça ekle |

> Spotify endpoint'leri için kullanıcının Spotify hesabını `/auth/spotify` ile bağlamış olması gerekir.

### Health

| Method | Endpoint | Açıklama |
|--------|----------|----------|
| `GET` | `/health` | API sağlık kontrolü |

---

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

**Playlist üret (ana akış):**
```bash
curl -X POST http://localhost:8080/api/v1/recommendations/generate \
  -H "Authorization: Bearer JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"text": "Bugün bahar gibiyim, biraz dansa ihtiyacım var", "limit": 20}'
```

**Spotify'da playlist oluştur:**
```bash
curl -X POST http://localhost:8080/api/v1/spotify/playlist \
  -H "Authorization: Bearer JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Sabah Enerjisi",
    "description": "Aria tarafından oluşturuldu",
    "public": false,
    "track_uris": ["spotify:track:3n3Ppam7vgaVa1iaRUc9Lp"]
  }'
```

**Google ile giriş:**
Tarayıcıda: `http://localhost:8080/api/v1/auth/google`

**Spotify ile giriş:**
Tarayıcıda: `http://127.0.0.1:8080/api/v1/auth/spotify`

---

## 🔐 OAuth 2.0 Entegrasyonu

### Google OAuth 2.0

**Gerekli env:**
```
GOOGLE_CLIENT_ID=...
GOOGLE_CLIENT_SECRET=...
GOOGLE_REDIRECT_URL=http://localhost:8080/api/v1/auth/google/callback
```

### Spotify OAuth 2.0

Spotify `access_token` ve `refresh_token` veritabanına kaydedilir. `SpotifyTokenManager` token süresi dolduğunda otomatik olarak yeniler.

**Gerekli env:**
```
SPOTIFY_CLIENT_ID=...
SPOTIFY_CLIENT_SECRET=...
SPOTIFY_REDIRECT_URL=http://127.0.0.1:8080/api/v1/auth/spotify/callback
```

> **Not:** Spotify `localhost` yerine `127.0.0.1` gerektirir.

**Spotify OAuth Scopes:**
`user-read-email`, `user-read-private`, `playlist-modify-public`, `playlist-modify-private`, `user-read-recently-played`, `user-top-read`

### OAuth Kullanıcı Eşleştirme Mantığı

1. Provider ID (google_id / spotify_id) ile DB araması
2. Bulunursa → mevcut kullanıcı döner + token'lar güncellenir
3. Bulunamazsa → email ile arama
4. Email varsa → provider ID hesaba bağlanır
5. Email de yoksa → yeni kullanıcı oluşturulur

---

## 🧠 Orchestrator Akışı — `POST /api/v1/recommendations/generate`

```
Frontend → POST /recommendations/generate
              │
              ├─ AuthMiddleware (JWT doğrulama)
              ├─ RateLimitMiddleware (Redis: 5 istek/dk)
              │
              ▼
    pipeline.Service.GeneratePlaylist
              │
    (a) mood.CreateRawMood        → DB: 'pending' Mood kaydı
    (b) aiclient.AnalyzeMood      → POST /analyze (Python servisi)
    (c) mood.UpdateAnalysis       → DB: sentiment alanları, status='analyzed'
    (d) aiclient.GetRecommendations → POST /recommend (Python servisi)
    (e) recommendation.CreateFromAI → DB: Recommendation + Tracks (tek transaction)
    (f) GenerateResponse{Mood, Recommendation} döner
```

Tüm modüller arası iletişim **direkt Go fonksiyon çağrısıyla** yapılır. Sadece (b) ve (d) adımları dış HTTP istekleridir.

**Hata davranışı:**

| Adım | Hata olursa | Mood durumu |
|------|-------------|-------------|
| (a) | DB hatası, hiç kayıt kalmaz | — |
| (b) | AI yanıt vermedi | `failed` olarak işaretlenir |
| (c) | DB update başarısız | `pending` kalır |
| (d) | AI yanıt vermedi | `analyzed` kalır, öneri yazılmaz |
| (e) | DB transaction başarısız | `analyzed` kalır |

**HTTP yanıt kodları:**

| Kod | Anlam |
|-----|-------|
| `200 OK` | Başarılı — `{ mood, recommendation: { tracks: [...] } }` |
| `400 Bad Request` | Boş/aşırı uzun metin, geçersiz limit |
| `401 Unauthorized` | JWT yok veya geçersiz |
| `429 Too Many Requests` | Rate limit aşıldı (`Retry-After` header'ı ile) |
| `502 Bad Gateway` | AI servisi 4xx/5xx döndürdü |
| `503 Service Unavailable` | AI servisine ulaşılamadı |
| `500 Internal Server Error` | DB veya beklenmeyen hata |

---

## 🚦 Rate Limiting

`POST /api/v1/recommendations/generate` — Redis tabanlı fixed-window:

| Kapsam | Limit | Pencere |
|--------|-------|---------|
| Kullanıcı başına | **5 istek** | **1 dakika** |

**Redis anahtar şeması:** `ratelimit:generate:<userID>`

**Response header'ları:**

| Header | Anlam |
|--------|-------|
| `X-RateLimit-Limit` | Pencere başına izin (5) |
| `X-RateLimit-Remaining` | Kalan istek hakkı |
| `Retry-After` | (sadece 429'da) Sıfırlanmaya kalan saniye |

**Fail-open politikası:** Redis erişilemezse istek **bloklanmaz**, uyarı loglanır ve geçirilir.

---

## 🔌 Python AI Servisi Sözleşmesi

`pkg/aiclient/contract.go` — Python ekibinin Pydantic modellerini birebir uyumlu yazması gereken **tek doğru kaynak**.

| Endpoint | Amaç | Request | Response |
|----------|------|---------|----------|
| `POST {AI_SERVICE_URL}/analyze` | Sentiment + duygu analizi | `AnalyzeRequest` | `AnalyzeResponse` |
| `POST {AI_SERVICE_URL}/recommend` | RAG tabanlı parça önerisi | `RecommendRequest` | `RecommendResponse` |

**Env değişkenleri:**
```
AI_SERVICE_URL=http://localhost:8000
AI_SERVICE_TIMEOUT_MS=15000
```

---

## 🗃️ Veritabanı Şeması

### Users

| Alan | Tip | Açıklama |
|------|-----|----------|
| id | uint PK | Primary key |
| email | string unique | Kullanıcı e-postası |
| password | string | Bcrypt hash (OAuth için opsiyonel) |
| google_id | string? | Google OAuth ID |
| spotify_id | string? | Spotify OAuth ID |
| spotify_access_token | string? | Spotify API token |
| spotify_refresh_token | string? | Spotify yenileme token'ı |
| created_at, updated_at | timestamp | — |

### Moods

| Alan | Tip | Açıklama |
|------|-----|----------|
| id | uint PK | Primary key |
| user_id | uint FK | Sahip kullanıcı (indexed) |
| raw_text | text | Kullanıcı girişi |
| sentiment_label | string | AI etiketi |
| dominant_emotion | string | En yüksek skorlu duygu |
| valence | float | [-1, +1] |
| arousal | float | [-1, +1] |
| energy | float | [0, 1] |
| emotion_scores | jsonb | Çok-etiketli duygu haritası |
| language | string | ISO 639-1 |
| ai_model_version | string | Analizi üreten model |
| processing_ms | int | AI süresi (ms) |
| status | string indexed | `pending` / `analyzed` / `failed` |

### Recommendations

| Alan | Tip | Açıklama |
|------|-----|----------|
| id | uint PK | Primary key |
| user_id | uint FK | Sahip kullanıcı |
| mood_id | uint FK | İlişkili Mood |
| ai_model_version | string | RAG model sürümü |
| rag_context | text | LLM bağlamı (debug/explainability) |
| processing_ms | int | AI süresi (ms) |
| status | string indexed | `pending` / `ready` / `failed` |

### RecommendedTracks

| Alan | Tip | Açıklama |
|------|-----|----------|
| id | uint PK | Primary key |
| recommendation_id | uint FK | Parent (CASCADE delete) |
| spotify_track_id | string? | Spotify katalog ID'si |
| title, artist, album | string | Parça meta verisi |
| preview_url, external_url | string | Spotify linkleri |
| duration_ms | int | Parça süresi (ms) |
| position | int | Sıra (0'dan başlar) |
| relevance_score | float | [0, 1] AI güven skoru |
| reason | text | "Neden bu şarkı?" açıklaması |

> Composite unique index: `(recommendation_id, position)` — aynı küme içinde çakışma engellenir.

---

## 🧪 Testler

Test dosyaları `Test/` klasöründe bulunur. Harici framework kullanılmaz, yalnızca Go standart kütüphanesi.

```bash
go test ./Test/... -v
```

| Dosya | Kapsam | Test Sayısı |
|-------|--------|-------------|
| `auth_jwt_test.go` | JWT üretim/doğrulama, bcrypt, blacklist | 9 |
| `mood_recommendation_test.go` | Model sabitleri, DTO'lar, sentinel hatalar | 11 |
| `pipeline_service_test.go` | Mock AI client, hata kategorileri, sözleşme alanları | 11 |
| `spotify_module_test.go` | Track/Artist/Playlist yapıları | 9 |

**Toplam: 40 test — tümü veritabanı/ağ bağlantısı gerektirmez.**

---

## 🌱 Seeder

Uygulama her başlatılışında otomatik olarak 10 örnek kullanıcı oluşturur (idempotent):

| Email | Şifre |
|-------|-------|
| alice@example.com | password123 |
| bob@example.com | password123 |
| charlie@example.com | password123 |
| ... | ... |

---

## 🔑 Ortam Değişkenleri

```env
# Sunucu
PORT=8080

# PostgreSQL
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres123
DB_NAME=music_curation

# Redis
REDIS_ADDR=localhost:6379

# JWT
JWT_SECRET=gizli-anahtar-buraya

# Google OAuth
GOOGLE_CLIENT_ID=...
GOOGLE_CLIENT_SECRET=...
GOOGLE_REDIRECT_URL=http://localhost:8080/api/v1/auth/google/callback

# Spotify OAuth
SPOTIFY_CLIENT_ID=...
SPOTIFY_CLIENT_SECRET=...
SPOTIFY_REDIRECT_URL=http://127.0.0.1:8080/api/v1/auth/spotify/callback

# Python AI Servisi
AI_SERVICE_URL=http://localhost:8000
AI_SERVICE_TIMEOUT_MS=15000
```

---

## 🔜 Yol Haritası

- [x] Proje iskeleti & Auth (signup, login, me, refresh, logout)
- [x] Google OAuth 2.0 entegrasyonu
- [x] Spotify OAuth 2.0 (Access & Refresh Token yönetimi)
- [x] Mood + Recommendation modülleri (GORM + jsonb)
- [x] AI/RAG client + Python servisi sözleşmesi (`pkg/aiclient`)
- [x] Orchestrator pipeline (`POST /recommendations/generate`)
- [x] Redis bağlantısı + per-user rate limiting (5/dk, fail-open)
- [x] Spotify API entegrasyonu (history, top tracks/artists, playlist oluşturma)
- [x] Test suite (`Test/` — 40 test, DB gerektirmez)
- [ ] Öneri sonuçlarının Redis ile cache'lenmesi
- [ ] Asenkron pipeline (job queue + status polling)

---

## 👥 Ekip & Sorumluluklar

| Modül | Geliştirici | Durum |
|-------|-------------|-------|
| Auth (signup/login/OAuth/refresh/logout) | Ortak | ✅ Tamamlandı |
| User, Mood, Recommendation modülleri | Bunyamin | ✅ Tamamlandı |
| AI client & Pipeline orchestrator | Bunyamin | ✅ Tamamlandı |
| Redis cache & Rate limiter | Bunyamin | ✅ Tamamlandı |
| Spotify API entegrasyonu (history/top/playlist) | Takım Arkadaşı | ✅ Tamamlandı |
| Python FastAPI AI/RAG servisi | Takım Arkadaşı | 🔜 Geliyor |
| Next.js Frontend | Takım Arkadaşı | 🔜 Geliyor |

---

## 📄 Lisans

Bu proje özel kullanım içindir (üniversite projesi).