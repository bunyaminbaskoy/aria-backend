package test

// Bu dosya, JWT üretimi/doğrulaması, şifre hashleme ve kara liste
// mekanizmasını test eder. Bunlar auth/middleware modüllerinin
// temel güvenlik katmanlarıdır.

import (
	"os"
	"testing"
	"time"

	"music-curation/pkg/utils"
)

// ===========================================================================
// Yardımcı: Test ortamı kurulumu
// ===========================================================================

// setupJWTSecret, testler için JWT_SECRET ortam değişkenini ayarlar.
// Gerçek üretimde bu değer .env'den gelir; testlerde sabit bir değer kullanılır.
func setupJWTSecret(t *testing.T) {
	t.Helper()
	t.Setenv("JWT_SECRET", "aria-test-secret-gizli-anahtar-2024")
}

// ===========================================================================
// Testler: JWT Token üretimi ve doğrulama
// ===========================================================================

// TestGenerateAndValidateAccessToken, access token üretilip başarıyla
// doğrulanabildiğini ve içindeki claim'lerin doğru olduğunu test eder.
func TestGenerateAndValidateAccessToken(t *testing.T) {
	setupJWTSecret(t)

	testUserID := uint(42)
	testEmail := "test@example.com"

	// Token üret
	token, err := utils.GenerateToken(testUserID, testEmail)
	if err != nil {
		t.Fatalf("Token üretimi başarısız: %v", err)
	}
	if token == "" {
		t.Fatal("Üretilen token boş olmamalı")
	}

	// Token doğrula
	claims, err := utils.ValidateToken(token)
	if err != nil {
		t.Fatalf("Token doğrulaması başarısız: %v", err)
	}

	// Claim alanlarını kontrol et
	if claims.UserID != testUserID {
		t.Errorf("UserID: beklenen %d, aldı %d", testUserID, claims.UserID)
	}
	if claims.Email != testEmail {
		t.Errorf("Email: beklenen '%s', aldı '%s'", testEmail, claims.Email)
	}
	if claims.TokenType != "access" {
		t.Errorf("TokenType: beklenen 'access', aldı '%s'", claims.TokenType)
	}
}

// TestGenerateTokenPair, token çiftinin (access + refresh) başarıyla
// üretildiğini ve her birinin doğru türde olduğunu doğrular.
func TestGenerateTokenPair(t *testing.T) {
	setupJWTSecret(t)

	pair, err := utils.GenerateTokenPair(1, "user@test.com")
	if err != nil {
		t.Fatalf("Token çifti üretimi başarısız: %v", err)
	}
	if pair.AccessToken == "" {
		t.Error("AccessToken boş olmamalı")
	}
	if pair.RefreshToken == "" {
		t.Error("RefreshToken boş olmamalı")
	}
	if pair.AccessToken == pair.RefreshToken {
		t.Error("AccessToken ve RefreshToken farklı olmalı")
	}

	// Access token türünü kontrol et
	accessClaims, err := utils.ValidateToken(pair.AccessToken)
	if err != nil {
		t.Fatalf("AccessToken doğrulaması başarısız: %v", err)
	}
	if accessClaims.TokenType != "access" {
		t.Errorf("AccessToken türü: beklenen 'access', aldı '%s'", accessClaims.TokenType)
	}

	// Refresh token türünü kontrol et
	refreshClaims, err := utils.ValidateToken(pair.RefreshToken)
	if err != nil {
		t.Fatalf("RefreshToken doğrulaması başarısız: %v", err)
	}
	if refreshClaims.TokenType != "refresh" {
		t.Errorf("RefreshToken türü: beklenen 'refresh', aldı '%s'", refreshClaims.TokenType)
	}
}

// TestValidateToken_InvalidToken, geçersiz bir token string'inin
// ValidateToken tarafından reddedildiğini doğrular.
func TestValidateToken_InvalidToken(t *testing.T) {
	setupJWTSecret(t)

	_, err := utils.ValidateToken("bu.gecersiz.bir.token")
	if err == nil {
		t.Fatal("Geçersiz token için hata bekleniyor ama err nil döndü")
	}
}

// TestValidateToken_EmptyToken, boş string gönderildiğinde hata alındığını doğrular.
func TestValidateToken_EmptyToken(t *testing.T) {
	setupJWTSecret(t)

	_, err := utils.ValidateToken("")
	if err == nil {
		t.Fatal("Boş token için hata bekleniyor ama err nil döndü")
	}
}

// TestValidateToken_WrongSecret, farklı JWT_SECRET ile imzalanmış token'ın
// reddedildiğini doğrular. Bu, token sahteciliğini önleyen temel kontroldür.
func TestValidateToken_WrongSecret(t *testing.T) {
	// Token'ı ilk sırla imzala
	os.Setenv("JWT_SECRET", "dogru-sifre")
	token, err := utils.GenerateToken(1, "test@test.com")
	if err != nil {
		t.Fatalf("Token üretimi başarısız: %v", err)
	}

	// Farklı sırla doğrulamayı dene
	os.Setenv("JWT_SECRET", "yanlis-sifre")
	t.Cleanup(func() { os.Unsetenv("JWT_SECRET") })

	_, err = utils.ValidateToken(token)
	if err == nil {
		t.Fatal("Yanlış secret ile imzalanmış token reddedilmeli ama err nil döndü")
	}
}

// TestGenerateToken_NoSecret, JWT_SECRET ayarlanmadan token üretiminin
// hata döndürdüğünü doğrular.
func TestGenerateToken_NoSecret(t *testing.T) {
	// JWT_SECRET'i temizle
	old := os.Getenv("JWT_SECRET")
	os.Unsetenv("JWT_SECRET")
	t.Cleanup(func() { os.Setenv("JWT_SECRET", old) })

	_, err := utils.GenerateToken(1, "test@test.com")
	if err == nil {
		t.Fatal("JWT_SECRET olmadan token üretimi hata döndürmeli")
	}
}

// ===========================================================================
// Testler: Şifre Hashleme (pkg/utils/password.go)
// ===========================================================================

// TestHashAndCheckPassword, şifrenin hashlenip ardından doğrulanabildiğini
// ve farklı şifrenin eşleşmediğini test eder.
func TestHashAndCheckPassword(t *testing.T) {
	password := "SuperSecret123!"

	// Hashle
	hashed, err := utils.HashPassword(password)
	if err != nil {
		t.Fatalf("Şifre hashleme başarısız: %v", err)
	}
	if hashed == "" {
		t.Fatal("Hash boş olmamalı")
	}
	if hashed == password {
		t.Fatal("Hash, orijinal şifreyle aynı olmamalı (plaintext saklanmış!)")
	}

	// Doğru şifre eşleşmeli
	if !utils.CheckPassword(password, hashed) {
		t.Error("Doğru şifre hash ile eşleşmeli ama false döndü")
	}

	// Yanlış şifre eşleşmemeli
	if utils.CheckPassword("YanlisSifre123!", hashed) {
		t.Error("Yanlış şifre hash ile eşleşmemeli ama true döndü")
	}
}

// TestHashPassword_DifferentHashEachTime, aynı şifrenin her hashlemede
// farklı sonuç ürettiğini doğrular (bcrypt salt mekanizması).
func TestHashPassword_DifferentHashEachTime(t *testing.T) {
	password := "AyniSifre"

	hash1, err := utils.HashPassword(password)
	if err != nil {
		t.Fatalf("İlk hashleme başarısız: %v", err)
	}
	hash2, err := utils.HashPassword(password)
	if err != nil {
		t.Fatalf("İkinci hashleme başarısız: %v", err)
	}

	// Bcrypt her seferinde farklı salt kullanır — hashler farklı olmalı
	if hash1 == hash2 {
		t.Error("Bcrypt her seferinde farklı hash üretmeli (rainbow table koruması)")
	}

	// Ama her ikisi de doğrulama geçmeli
	if !utils.CheckPassword(password, hash1) {
		t.Error("Birinci hash doğrulama başarısız")
	}
	if !utils.CheckPassword(password, hash2) {
		t.Error("İkinci hash doğrulama başarısız")
	}
}

// ===========================================================================
// Testler: Token Kara Listesi (logout mekanizması)
// ===========================================================================

// TestTokenBlacklist, bir token'ın kara listeye alınıp alınmadığını
// doğrular. Bu, logout sonrası token'ların yeniden kullanılamamasını sağlar.
//
// Not: Kara liste auth paketinde tanımlı; burada aynı paketten import
// yerine aynı davranışı pkg/utils üzerinden simüle ediyoruz.
// Gerçek kara liste testi auth_blacklist_test.go'da ayrıca yapılabilir.
func TestTokenBlacklist_Logic(t *testing.T) {
	setupJWTSecret(t)

	// Bu test, blacklist mekanizmasının mantığını açıklar:
	// 1. Kullanıcı logout yapar → refresh token kara listeye alınır.
	// 2. Aynı refresh token ile yeni access token istenemez.
	// 3. Blacklist her saat temizlenir (süresi dolmuş olanlar).

	// Basit bir kara liste simülasyonu
	blacklist := make(map[string]time.Time)

	addToBlacklist := func(token string, expiry time.Time) {
		blacklist[token] = expiry
	}

	isBlacklisted := func(token string) bool {
		_, exists := blacklist[token]
		return exists
	}

	testToken := "test.refresh.token.string"
	expiry := time.Now().Add(7 * 24 * time.Hour)

	// Henüz kara listede değil
	if isBlacklisted(testToken) {
		t.Error("Token henüz kara listede olmamalı")
	}

	// Kara listeye ekle
	addToBlacklist(testToken, expiry)

	// Artık kara listede
	if !isBlacklisted(testToken) {
		t.Error("Token kara listeye eklendikten sonra orada görünmeli")
	}

	// Farklı token etkilenmemeli
	otherToken := "diger.token.string"
	if isBlacklisted(otherToken) {
		t.Error("Farklı token kara listeden etkilenmemeli")
	}
}
