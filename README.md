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
| Redis | Cache & oturum yönetimi (yakında) |

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
│   │   └── auth.go              # JWT doğrulama middleware'i
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

**4. Bağımlılıkları indir:**

```bash
go mod tidy
```

**5. Sunucuyu başlat:**

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
- [ ] Spotify API ile playlist oluşturma
- [ ] Redis cache entegrasyonu
- [ ] RAG tabanlı müzik öneri sistemi

## 👥 Ekip

| Görev | Durum |
|-------|-------|
| Proje iskeleti & Auth | ✅ Tamamlandı |
| Google/Spotify OAuth | ✅ Tamamlandı |
| Spotify API entegrasyonu | 🔜 Geliyor |

## 📄 Lisans

Bu proje özel kullanım içindir.