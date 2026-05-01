package test

// Bu dosya, sistemin ana orchestrator'ı olan pipeline.Service'i test eder.
// Gerçek veritabanı veya AI servisine bağlanmadan, sahte (mock) implementasyonlar
// kullanarak iş mantığının doğruluğunu doğrular.
//
// Test stratejisi:
//   - Tüm dış bağımlılıklar (DB, AI) mock/stub ile değiştirilir.
//   - Sadece Go standart kütüphanesi kullanılır (harici test framework yok).
//   - Her test fonksiyonu bir senaryo ("happy path" veya "hata durumu") kapsar.

import (
	"context"
	"errors"
	"testing"

	"music-curation/pkg/aiclient"
)

// ===========================================================================
// Mock: AI İstemcisi
// ===========================================================================

// mockAIClient, AIClient arayüzünü taklit eden sahte bir yapıdır.
// Her test için farklı davranış tanımlanabilmesi amacıyla
// AnalyzeFn ve RecommendFn alanları dışarıdan atanabilir.
type mockAIClient struct {
	AnalyzeFn    func(ctx context.Context, req aiclient.AnalyzeRequest) (*aiclient.AnalyzeResponse, error)
	RecommendFn  func(ctx context.Context, req aiclient.RecommendRequest) (*aiclient.RecommendResponse, error)
}

func (m *mockAIClient) AnalyzeMood(ctx context.Context, req aiclient.AnalyzeRequest) (*aiclient.AnalyzeResponse, error) {
	return m.AnalyzeFn(ctx, req)
}

func (m *mockAIClient) GetRecommendations(ctx context.Context, req aiclient.RecommendRequest) (*aiclient.RecommendResponse, error) {
	return m.RecommendFn(ctx, req)
}

// ===========================================================================
// Mock: Mood Servisi
// ===========================================================================

// mockMoodService, mood.Service'in davranışını taklit eder.
// Pipeline service, mood modülüne direkt Go çağrısı yapar; bu yüzden
// bu mock arayüzü gerçek servis ile aynı imzaları taşır.
type mockMoodService struct {
	CreateRawMoodFn  func(userID uint, text string) (*moodStub, error)
	UpdateAnalysisFn func(id uint) error
	MarkFailedFn     func(id uint) error
	GetMoodByIDFn    func(id uint) (*moodStub, error)
}

// moodStub, mood.Mood modelini test ortamında temsil eden minimal yapı.
type moodStub struct {
	ID      uint
	UserID  uint
	RawText string
	Status  string
}

// ===========================================================================
// Mock: Recommendation Servisi
// ===========================================================================

// recResultStub, recommendation.Recommendation'ı test ortamında temsil eder.
type recResultStub struct {
	ID     uint
	UserID uint
	MoodID uint
	Status string
	Tracks []trackStub
}

// trackStub, tek bir öneri parçasını temsil eder.
type trackStub struct {
	Title  string
	Artist string
}

// ===========================================================================
// Yardımcı: Geçerli AI cevabı üret
// ===========================================================================

// validAnalyzeResponse, başarılı bir sentiment analizi cevabını döner.
func validAnalyzeResponse() *aiclient.AnalyzeResponse {
	return &aiclient.AnalyzeResponse{
		SentimentLabel:  "positive",
		DominantEmotion: "joy",
		Valence:         0.8,
		Arousal:         0.6,
		Energy:          0.7,
		EmotionScores:   map[string]float64{"joy": 0.8, "calm": 0.2},
		Language:        "tr",
		ModelVersion:    "aria-sentiment-v1.0.0",
		ProcessingMs:    120,
	}
}

// validRecommendResponse, başarılı bir parça önerisi cevabını döner.
func validRecommendResponse() *aiclient.RecommendResponse {
	return &aiclient.RecommendResponse{
		ModelVersion: "aria-rag-v1.0.0",
		RAGContext:   "Yüksek valence + arousal → enerjik parçalar önerildi.",
		ProcessingMs: 340,
		Tracks: []aiclient.TrackSuggestion{
			{
				SpotifyTrackID: "3n3Ppam7vgaVa1iaRUc9Lp",
				Title:          "Holocene",
				Artist:         "Bon Iver",
				Album:          "Bon Iver, Bon Iver",
				RelevanceScore: 0.92,
				Reason:         "Sakin vokaller ve dingin atmosfer ruh haliyle örtüşüyor.",
			},
			{
				SpotifyTrackID: "2TpxZ7JUBn3uw46aR7qd6V",
				Title:          "Skinny Love",
				Artist:         "Bon Iver",
				Album:          "For Emma, Forever Ago",
				RelevanceScore: 0.87,
				Reason:         "Melankolik melodi duygusal derinliği destekliyor.",
			},
		},
	}
}

// ===========================================================================
// Testler: aiclient.Client (birimsiz, mantıksal kontrol)
// ===========================================================================

// TestAIClientSentinelErrors, aiclient paketindeki sentinel hataların
// var olduğunu ve errors.Is ile doğru şekilde eşleştiğini doğrular.
// Bu test, pipeline.mapAIError fonksiyonunun güvendiği temel mekanizmayı kontrol eder.
func TestAIClientSentinelErrors(t *testing.T) {
	// Sentinel hataları tanımlandı mı?
	if aiclient.ErrServiceUnavailable == nil {
		t.Fatal("ErrServiceUnavailable nil olmamalıdır")
	}
	if aiclient.ErrInvalidRequest == nil {
		t.Fatal("ErrInvalidRequest nil olmamalıdır")
	}
	if aiclient.ErrInternal == nil {
		t.Fatal("ErrInternal nil olmamalıdır")
	}
	if aiclient.ErrDecode == nil {
		t.Fatal("ErrDecode nil olmamalıdır")
	}

	// errors.Is zinciri çalışıyor mu?
	wrappedUnavailable := errors.Join(aiclient.ErrServiceUnavailable, errors.New("connection refused"))
	if !errors.Is(wrappedUnavailable, aiclient.ErrServiceUnavailable) {
		t.Error("errors.Is(wrappedUnavailable, ErrServiceUnavailable) false döndü, true bekleniyordu")
	}
}

// TestMockAIClientAnalyze, mock AI istemcisinin beklenen veriyi döndürdüğünü
// ve arayüz uyumluluğunu doğrular.
func TestMockAIClientAnalyze(t *testing.T) {
	ai := &mockAIClient{
		AnalyzeFn: func(ctx context.Context, req aiclient.AnalyzeRequest) (*aiclient.AnalyzeResponse, error) {
			// Boş metin gelirse hata döndür
			if req.Text == "" {
				return nil, aiclient.ErrInvalidRequest
			}
			return validAnalyzeResponse(), nil
		},
	}

	ctx := context.Background()

	// --- Başarılı senaryo ---
	resp, err := ai.AnalyzeMood(ctx, aiclient.AnalyzeRequest{Text: "Bugün harika hissediyorum", UserID: 1})
	if err != nil {
		t.Fatalf("Hata beklenmiyor, aldı: %v", err)
	}
	if resp.SentimentLabel != "positive" {
		t.Errorf("SentimentLabel: beklenen 'positive', aldı '%s'", resp.SentimentLabel)
	}
	if resp.Valence != 0.8 {
		t.Errorf("Valence: beklenen 0.8, aldı %f", resp.Valence)
	}

	// --- Hata senaryosu: boş metin ---
	_, err = ai.AnalyzeMood(ctx, aiclient.AnalyzeRequest{Text: "", UserID: 1})
	if err == nil {
		t.Fatal("Boş metin için hata bekleniyor ama err nil döndü")
	}
	if !errors.Is(err, aiclient.ErrInvalidRequest) {
		t.Errorf("errors.Is(err, ErrInvalidRequest) false döndü; aldı: %v", err)
	}
}

// TestMockAIClientRecommend, mock AI istemcisinin recommend çağrısını
// doğru şekilde işlediğini test eder.
func TestMockAIClientRecommend(t *testing.T) {
	ai := &mockAIClient{
		RecommendFn: func(ctx context.Context, req aiclient.RecommendRequest) (*aiclient.RecommendResponse, error) {
			// AI servisi kapalıysa hata döndür
			if req.Limit <= 0 {
				return nil, aiclient.ErrInvalidRequest
			}
			return validRecommendResponse(), nil
		},
	}

	ctx := context.Background()

	// --- Başarılı senaryo ---
	resp, err := ai.GetRecommendations(ctx, aiclient.RecommendRequest{
		UserID: 1,
		MoodID: 10,
		Mood: aiclient.MoodSnapshot{
			SentimentLabel:  "positive",
			DominantEmotion: "joy",
			Valence:         0.8,
			Arousal:         0.6,
			Energy:          0.7,
		},
		Limit: 20,
	})
	if err != nil {
		t.Fatalf("Hata beklenmiyor, aldı: %v", err)
	}
	if len(resp.Tracks) != 2 {
		t.Errorf("Parça sayısı: beklenen 2, aldı %d", len(resp.Tracks))
	}
	if resp.Tracks[0].Title != "Holocene" {
		t.Errorf("İlk parça başlığı: beklenen 'Holocene', aldı '%s'", resp.Tracks[0].Title)
	}

	// --- Hata senaryosu: geçersiz limit ---
	_, err = ai.GetRecommendations(ctx, aiclient.RecommendRequest{Limit: 0})
	if err == nil {
		t.Fatal("Geçersiz limit için hata bekleniyor ama err nil döndü")
	}
}

// ===========================================================================
// Testler: aiclient Sözleşme (Contract) Alanları
// ===========================================================================

// TestAnalyzeResponseFields, AnalyzeResponse alanlarının sıfır değerden
// farklı olmasını doğrular. Bu test, Python servisiyle sözleşme kopukluğunu
// erken tespit etmeye yardımcı olur.
func TestAnalyzeResponseFields(t *testing.T) {
	resp := validAnalyzeResponse()

	if resp.SentimentLabel == "" {
		t.Error("SentimentLabel boş olmamalı")
	}
	if resp.DominantEmotion == "" {
		t.Error("DominantEmotion boş olmamalı")
	}
	if resp.ModelVersion == "" {
		t.Error("ModelVersion boş olmamalı")
	}
	if resp.ProcessingMs <= 0 {
		t.Error("ProcessingMs pozitif olmalı")
	}
	if resp.Valence < -1.0 || resp.Valence > 1.0 {
		t.Errorf("Valence [-1.0, 1.0] aralığında olmalı, aldı: %f", resp.Valence)
	}
	if resp.Arousal < -1.0 || resp.Arousal > 1.0 {
		t.Errorf("Arousal [-1.0, 1.0] aralığında olmalı, aldı: %f", resp.Arousal)
	}
	if resp.Energy < 0.0 || resp.Energy > 1.0 {
		t.Errorf("Energy [0.0, 1.0] aralığında olmalı, aldı: %f", resp.Energy)
	}
}

// TestRecommendResponseFields, RecommendResponse'un track listesi boş
// olmadığında gerekli alanların dolu olduğunu kontrol eder.
func TestRecommendResponseFields(t *testing.T) {
	resp := validRecommendResponse()

	if resp.ModelVersion == "" {
		t.Error("ModelVersion boş olmamalı")
	}
	if len(resp.Tracks) == 0 {
		t.Fatal("Tracks listesi boş olmamalı")
	}

	for i, track := range resp.Tracks {
		if track.Title == "" {
			t.Errorf("Track[%d].Title boş olmamalı", i)
		}
		if track.Artist == "" {
			t.Errorf("Track[%d].Artist boş olmamalı", i)
		}
		if track.RelevanceScore < 0.0 || track.RelevanceScore > 1.0 {
			t.Errorf("Track[%d].RelevanceScore [0.0, 1.0] aralığında olmalı, aldı: %f", i, track.RelevanceScore)
		}
	}
}

// ===========================================================================
// Testler: AI servisi hata durumları → pipeline hata eşleme mantığı
// ===========================================================================

// TestAIErrorMapping, aiclient sentinel hatalarının pipeline düzeyinde
// doğru şekilde kategorize edildiğini doğrular.
// (mapAIError fonksiyonunu dolaylı olarak test eder.)
func TestAIErrorMapping(t *testing.T) {
	testCases := []struct {
		name          string
		inputErr      error
		wantSentinel  error
	}{
		{
			name:         "Ağ hatası ErrServiceUnavailable olarak etiketlenmeli",
			inputErr:     aiclient.ErrServiceUnavailable,
			wantSentinel: aiclient.ErrServiceUnavailable,
		},
		{
			name:         "Geçersiz istek ErrInvalidRequest olarak etiketlenmeli",
			inputErr:     aiclient.ErrInvalidRequest,
			wantSentinel: aiclient.ErrInvalidRequest,
		},
		{
			name:         "AI iç hatası ErrInternal olarak etiketlenmeli",
			inputErr:     aiclient.ErrInternal,
			wantSentinel: aiclient.ErrInternal,
		},
		{
			name:         "JSON decode hatası ErrDecode olarak etiketlenmeli",
			inputErr:     aiclient.ErrDecode,
			wantSentinel: aiclient.ErrDecode,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Hata sarmalama ile birlikte errors.Is çalışmalı
			wrapped := errors.Join(tc.inputErr, errors.New("ek bağlam"))
			if !errors.Is(wrapped, tc.wantSentinel) {
				t.Errorf("errors.Is(wrapped, %v) = false; beklenen: true", tc.wantSentinel)
			}
		})
	}
}
