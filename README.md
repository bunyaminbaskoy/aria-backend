# 🎵 Aria — RAG Tabanlı Müzik Kürasyonu Sistemi

Modular Monolith mimarisinde Go (Golang) ile geliştirilmiş bir müzik kürasyonu backend API'si.

## 🏗️ Teknoloji Stack

| Teknoloji | Kullanım |
|-----------|----------|
| **Go 1.26** | Ana programlama dili |
| **Gin** | HTTP web framework |
| **GORM** | ORM (Object-Relational Mapping) |
| **PostgreSQL** | Ana veritabanı |
| **JWT** | Token tabanlı kimlik doğrulama |
| **Bcrypt** | Şifre hashleme |
| **Redis** | Cache & oturum yönetimi *(yakında)* |

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
│   │   └── postgres.go          # PostgreSQL bağlantı konfigürasyonu
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

- [Go 1.22+](https://go.dev/dl/)
- [PostgreSQL](https://www.postgresql.org/download/)

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
`.env` dosyasını düzenleyip kendi PostgreSQL bilgilerini gir.

**3. Veritabanını oluştur:**
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

### Auth (Public)

| Method | Endpoint | Açıklama |
|--------|----------|----------|
| `POST` | `/api/v1/auth/signup` | Yeni kullanıcı kaydı |
| `POST` | `/api/v1/auth/login` | Giriş yap, JWT al |
| `GET` | `/api/v1/auth/me` | Mevcut kullanıcıyı getir 🔒 |

### Users (Protected — JWT gerekli 🔒)

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

## 🗃️ Veritabanı Şeması

### Users Tablosu

| Alan | Tip | Açıklama |
|------|-----|----------|
| `id` | `uint` (PK) | Otomatik artan primary key |
| `email` | `string` (Unique) | Kullanıcı e-postası |
| `password` | `string` | Bcrypt ile hashlenmiş şifre |
| `google_id` | `string?` | Google OAuth ID (nullable) |
| `spotify_id` | `string?` | Spotify OAuth ID (nullable) |
| `created_at` | `timestamp` | Oluşturulma tarihi |
| `updated_at` | `timestamp` | Güncellenme tarihi |

## 🌱 Seeder

Uygulama ilk başlatıldığında otomatik olarak 10 örnek kullanıcı oluşturur:

| Email | Şifre |
|-------|-------|
| alice@example.com | password123 |
| bob@example.com | password123 |
| charlie@example.com | password123 |
| ... | ... |

> Seeder idempotent'tır — tekrar çalıştırıldığında mevcut kullanıcıları tekrar eklemez.

## 🔜 Yol Haritası

- [ ] Google OAuth 2.0 entegrasyonu
- [ ] Spotify OAuth 2.0 (Access & Refresh Token yönetimi)
- [ ] Spotify API ile playlist oluşturma
- [ ] Redis cache entegrasyonu
- [ ] RAG tabanlı müzik öneri sistemi

## 👥 Ekip

| Görev | Durum |
|-------|-------|
| Proje iskeleti & Auth | ✅ Tamamlandı |
| Google/Spotify OAuth | 🔜 Geliyor |
| Spotify API entegrasyonu | 🔜 Geliyor |

## 📄 Lisans

Bu proje özel kullanım içindir.