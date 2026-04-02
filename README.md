# aria-backend

Music Curation API — Modüler Monolit mimaride Go (Gin) backend uygulaması.

## Tech Stack

- **Go 1.26** + **Gin** (HTTP framework)
- **GORM** + **PostgreSQL** (ORM & veritabanı)
- **JWT** (golang-jwt/v5) — kimlik doğrulama
- **OAuth 2.0** (golang.org/x/oauth2) — Google & Spotify entegrasyonu
- **Docker** — PostgreSQL container

## Kurulum

### 1. PostgreSQL (Docker ile)

```bash
docker run -d \
  --name aria-postgres \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=postgres123 \
  -e POSTGRES_DB=music_curation \
  -p 5432:5432 \
  postgres:16-alpine
```

### 2. Environment Variables

`.env.example` dosyasını `.env` olarak kopyala ve değerleri doldur:

```bash
cp .env.example .env
```

### 3. Uygulamayı Başlat

```bash
go run ./cmd/api/main.go
```

Uygulama `http://localhost:8080` üzerinde çalışacaktır.

## API Endpoints

### Auth (Kimlik Doğrulama)

| Method | Path | Auth | Açıklama |
|--------|------|------|----------|
| `POST` | `/api/v1/auth/signup` | ❌ | Email + şifre ile kayıt |
| `POST` | `/api/v1/auth/login` | ❌ | Email + şifre ile giriş |
| `GET` | `/api/v1/auth/me` | ✅ JWT | Aktif kullanıcı bilgisi |
| `GET` | `/api/v1/auth/google` | ❌ | Google OAuth — giriş ekranına yönlendirir |
| `GET` | `/api/v1/auth/google/callback` | ❌ | Google OAuth callback — JWT döner |
| `GET` | `/api/v1/auth/spotify` | ❌ | Spotify OAuth — giriş ekranına yönlendirir |
| `GET` | `/api/v1/auth/spotify/callback` | ❌ | Spotify OAuth callback — JWT döner |

### Users (Kullanıcılar)

| Method | Path | Auth | Açıklama |
|--------|------|------|----------|
| `GET` | `/api/v1/users` | ✅ JWT | Tüm kullanıcıları listele |
| `GET` | `/api/v1/users/:id` | ✅ JWT | Tek kullanıcı getir |

### Health

| Method | Path | Açıklama |
|--------|------|----------|
| `GET` | `/health` | API durum kontrolü |

## OAuth Entegrasyonu (Yeni Eklenen)

### Google OAuth 2.0

- Kullanıcı `/api/v1/auth/google` adresine gider → Google consent ekranına yönlendirilir
- Giriş sonrası callback'e döner → Google'dan email ve google_id alınır
- DB'de kullanıcı varsa eşleştirilir, yoksa yeni oluşturulur
- JWT token döner

**Gerekli env değişkenleri:**
```
GOOGLE_CLIENT_ID=...
GOOGLE_CLIENT_SECRET=...
GOOGLE_REDIRECT_URL=http://localhost:8080/api/v1/auth/google/callback
```

### Spotify OAuth 2.0

- Kullanıcı `/api/v1/auth/spotify` adresine gider → Spotify authorize sayfasına yönlendirilir
- Giriş sonrası callback'e döner → Spotify'dan email, spotify_id, access_token, refresh_token alınır
- DB'de kullanıcı varsa eşleştirilir ve token'lar güncellenir, yoksa yeni oluşturulur
- JWT token döner
- Spotify `access_token` ve `refresh_token` DB'ye kaydedilir (ileride Spotify API çağrıları için)

**Gerekli env değişkenleri:**
```
SPOTIFY_CLIENT_ID=...
SPOTIFY_CLIENT_SECRET=...
SPOTIFY_REDIRECT_URL=http://127.0.0.1:8080/api/v1/auth/spotify/callback
```

> **Not:** Spotify, localhost yerine `127.0.0.1` kullanılmasını gerektiriyor. Test ederken tarayıcıda da `127.0.0.1` kullanın.

**Spotify OAuth Scopes:**
- `user-read-email` — kullanıcı emaili
- `user-read-private` — kullanıcı profili
- `playlist-modify-public` — public playlist oluşturma
- `playlist-modify-private` — private playlist oluşturma

### OAuth Kullanıcı Eşleştirme Mantığı

1. Provider ID ile DB'de arama yapılır (google_id veya spotify_id)
2. Bulunursa → mevcut kullanıcı döner
3. Bulunamazsa → email ile arama yapılır
4. Email varsa → provider ID mevcut hesaba bağlanır
5. Email de yoksa → yeni kullanıcı oluşturulur

## Proje Yapısı

```
aria-backend/
├── cmd/api/main.go                    # Uygulama giriş noktası
├── internal/
│   ├── auth/
│   │   ├── dto.go                     # Request/Response DTO'ları
│   │   ├── handler.go                 # Signup, Login, Me handler'ları
│   │   ├── oauth_google.go            # Google OAuth config + handler
│   │   ├── oauth_spotify.go           # Spotify OAuth config + handler
│   │   └── routes.go                  # Auth route tanımları
│   ├── middleware/
│   │   └── auth.go                    # JWT auth middleware
│   ├── music/
│   │   └── model.go                   # Müzik modelleri (TODO)
│   ├── seeder/
│   │   └── seeder.go                  # Seed data
│   └── user/
│       ├── model.go                   # User modeli
│       ├── repository.go              # DB erişim katmanı
│       ├── service.go                 # Business logic
│       ├── handler.go                 # HTTP handler'lar
│       └── routes.go                  # User route tanımları
├── pkg/
│   ├── database/
│   │   └── postgres.go                # PostgreSQL bağlantısı
│   └── utils/
│       ├── jwt.go                     # JWT generate & validate
│       └── password.go                # bcrypt hash & check
├── go.mod
├── go.sum
├── .env.example
└── .gitignore
```

## User Modeli

```go
type User struct {
    ID                  uint
    Email               string   // Unique, zorunlu
    Password            string   // Opsiyonel (OAuth kullanıcıları için boş)
    GoogleID            *string  // Nullable — Google OAuth
    SpotifyID           *string  // Nullable — Spotify OAuth
    SpotifyAccessToken  *string  // Spotify API erişim token'ı
    SpotifyRefreshToken *string  // Spotify token yenileme
    CreatedAt           time.Time
    UpdatedAt           time.Time
}
```