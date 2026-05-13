package aiclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// Varsayılan konfigürasyon değerleri — env değişkenleri ayarlanmamışsa
// bu değerler kullanılır. Yerel geliştirme için Python FastAPI servisi
// genellikle 8000 portunda çalışır.
const (
	defaultBaseURL   = "http://localhost:8000"
	defaultTimeoutMs = 15000

	analyzePath   = "/api/v1/analyze"
	recommendPath = "/api/v1/recommend"
)

// Sentinel error'lar — pipeline ve handler katmanları, hatanın türüne
// göre farklı HTTP durum kodları döndürmek için errors.Is ile bunları
// kontrol eder. Hata tipini int'e (HTTP status) gömmek yerine semantic
// kategoriler kullanmak, üst katmanları AI servis detaylarından soyutlar.
var (
	// ErrServiceUnavailable, AI servisine ağ üzerinden ulaşılamadığında
	// (DNS, connection refused, timeout, vb.) döner. Pipeline bunu
	// 503'e çevirir.
	ErrServiceUnavailable = errors.New("AI servisine ulaşılamadı")

	// ErrInvalidRequest, AI servisi 4xx döndürdüğünde fırlatılır.
	// İstek gövdesinde bir sorun var demektir; backend tarafında
	// validate edilmesi gereken alanlar olabilir.
	ErrInvalidRequest = errors.New("AI servisine yapılan istek geçersiz")

	// ErrInternal, AI servisi 5xx döndürdüğünde fırlatılır.
	// Python tarafında model yüklenemedi, prompt başarısız vb.
	ErrInternal = errors.New("AI servisinde dahili hata oluştu")

	// ErrDecode, AI yanıtı JSON olarak çözümlenemediğinde döner.
	// Sözleşme (contract) ihlali anlamına gelir; loglanmalıdır.
	ErrDecode = errors.New("AI yanıtı çözümlenemedi")
)

// Client, Python AI/RAG servisi ile HTTP üzerinden konuşan tip-güvenli
// istemcidir. Tüm metotlar context.Context alır; üst katman (Gin
// request context'i) iptal edildiğinde HTTP isteği de iptal edilir.
//
// Goroutine güvenlidir — tek bir Client tüm uygulama için yeterlidir.
type Client struct {
	baseURL string
	http    *http.Client
}

// NewClient, env değişkenlerinden konfigürasyonu okuyarak yeni bir
// AI client örneği oluşturur. main.go içinde tek seferlik çağrılması
// hedeflenmiştir — http.Client'in kendi connection pool'u vardır,
// her istekte yeni client yaratmak performansı düşürür.
//
// Okunan env değişkenleri:
//
//	AI_SERVICE_URL         — taban URL (varsayılan: http://localhost:8000)
//	AI_SERVICE_TIMEOUT_MS  — istek başına timeout, milisaniye (varsayılan: 15000)
func NewClient() *Client {
	baseURL := strings.TrimSpace(os.Getenv("AI_SERVICE_URL"))
	if baseURL == "" {
		baseURL = defaultBaseURL
		log.Printf("⚠️  AI_SERVICE_URL bulunamadı, varsayılan kullanılıyor: %s", baseURL)
	}
	// Sondaki "/" karakterini at — path'leri eklerken çift slash olmasın.
	baseURL = strings.TrimRight(baseURL, "/")

	timeoutMs, _ := strconv.Atoi(os.Getenv("AI_SERVICE_TIMEOUT_MS"))
	if timeoutMs <= 0 {
		timeoutMs = defaultTimeoutMs
	}

	return &Client{
		baseURL: baseURL,
		http: &http.Client{
			Timeout: time.Duration(timeoutMs) * time.Millisecond,
		},
	}
}

// AnalyzeMood, kullanıcının ham ruh hali metnini Python servisine
// gönderir ve sentiment + duygu skorlarını döner.
//
// Hata durumunda *AnalyzeResponse nil döner; hata türünü öğrenmek için
// errors.Is(err, ErrServiceUnavailable) gibi kontroller yapılabilir.
func (c *Client) AnalyzeMood(ctx context.Context, req AnalyzeRequest) (*AnalyzeResponse, error) {
	var resp AnalyzeResponse
	if err := c.do(ctx, analyzePath, req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetRecommendations, sentiment analiz sonucundan yola çıkarak RAG
// tabanlı parça önerileri ister. Bu metot, AnalyzeMood'dan dönen
// sonuçların MoodSnapshot olarak gömülmesini bekler — böylece Python
// tarafı sentiment analizini tekrar yapmak zorunda kalmaz.
func (c *Client) GetRecommendations(ctx context.Context, req RecommendRequest) (*RecommendResponse, error) {
	var resp RecommendResponse
	if err := c.do(ctx, recommendPath, req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// do, AI servisine ortak POST + JSON gönderme/alma akışını çalıştıran
// iç metottur. Hem AnalyzeMood hem de GetRecommendations bunu kullanır.
//
// Akış:
//  1. İstek struct'ını JSON'a serileştir.
//  2. context'li HTTP request oluştur.
//  3. http.Client.Do ile gönder; ağ hatasında ErrServiceUnavailable.
//  4. Status koduna göre hata türünü belirle (4xx → ErrInvalidRequest,
//     5xx → ErrInternal).
//  5. 2xx ise gövdeyi out'a unmarshal et.
func (c *Client) do(ctx context.Context, path string, in, out interface{}) error {
	body, err := json.Marshal(in)
	if err != nil {
		// İstek struct'ı serileştirilemiyorsa bu bizim bug'ımızdır.
		return fmt.Errorf("istek serileştirilemedi: %w", err)
	}

	url := c.baseURL + path
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("HTTP isteği oluşturulamadı: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(httpReq)
	if err != nil {
		// Network error: connection refused, DNS hatası, timeout, vb.
		// context iptal edildiğinde de buraya düşer; aynı kategoride
		// ele alıyoruz çünkü her iki durumda da AI yanıt vermedi.
		return fmt.Errorf("%w: %v", ErrServiceUnavailable, err)
	}
	defer resp.Body.Close()

	respBody, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return fmt.Errorf("%w: gövde okunamadı: %v", ErrServiceUnavailable, readErr)
	}

	// Başarılı yanıt — gövdeyi out'a unmarshal et.
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		if err := json.Unmarshal(respBody, out); err != nil {
			return fmt.Errorf("%w: %v", ErrDecode, err)
		}
		return nil
	}

	// Hata yanıtı — contract'taki ErrorResponse'u parse etmeyi dene.
	// Parse edilemese bile status code'a göre kategori belirleyebiliriz.
	var apiErr ErrorResponse
	_ = json.Unmarshal(respBody, &apiErr)
	detail := apiErr.Message
	if detail == "" {
		detail = strings.TrimSpace(string(respBody))
	}

	switch {
	case resp.StatusCode >= 400 && resp.StatusCode < 500:
		return fmt.Errorf("%w: %s (HTTP %d)", ErrInvalidRequest, detail, resp.StatusCode)
	case resp.StatusCode >= 500:
		return fmt.Errorf("%w: %s (HTTP %d)", ErrInternal, detail, resp.StatusCode)
	default:
		return fmt.Errorf("AI servisinden beklenmeyen yanıt: HTTP %d, gövde=%s", resp.StatusCode, detail)
	}
}
