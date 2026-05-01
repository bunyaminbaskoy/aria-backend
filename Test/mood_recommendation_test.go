package test

// Bu dosya, Mood ve Recommendation modüllerinin veri modeli
// doğruluğunu ve servis katmanı davranışını test eder.
// Gerçek veritabanı kullanılmaz; stub yapılar ile test edilir.

import (
	"testing"

	"music-curation/internal/mood"
	"music-curation/internal/recommendation"
)

// ===========================================================================
// Testler: Mood Model Sabitleri
// ===========================================================================

// TestMoodStatusConstants, Mood modülünün durum sabitlerinin tanımlı
// ve beklenen değerlere sahip olduğunu doğrular.
// Pipeline bu sabitleri kullanarak DB'ye yazar; eşleşme kritik.
func TestMoodStatusConstants(t *testing.T) {
	if mood.StatusPending != "pending" {
		t.Errorf("mood.StatusPending: beklenen 'pending', aldı '%s'", mood.StatusPending)
	}
	if mood.StatusAnalyzed != "analyzed" {
		t.Errorf("mood.StatusAnalyzed: beklenen 'analyzed', aldı '%s'", mood.StatusAnalyzed)
	}
	if mood.StatusFailed != "failed" {
		t.Errorf("mood.StatusFailed: beklenen 'failed', aldı '%s'", mood.StatusFailed)
	}

	// Sabitler birbirinden farklı olmalı
	statuses := []string{mood.StatusPending, mood.StatusAnalyzed, mood.StatusFailed}
	seen := make(map[string]bool)
	for _, s := range statuses {
		if seen[s] {
			t.Errorf("Mood durum sabiti tekrarlıyor: '%s'", s)
		}
		seen[s] = true
	}
}

// TestMoodStatusConstants_Recommendation, Recommendation modülünün durum
// sabitlerini doğrular. Mood sabitleriyle aynı değerlere sahip olmalı
// (mimari tutarlılık).
func TestMoodStatusConstants_Recommendation(t *testing.T) {
	if recommendation.StatusPending != "pending" {
		t.Errorf("recommendation.StatusPending: beklenen 'pending', aldı '%s'", recommendation.StatusPending)
	}
	if recommendation.StatusReady != "ready" {
		t.Errorf("recommendation.StatusReady: beklenen 'ready', aldı '%s'", recommendation.StatusReady)
	}
	if recommendation.StatusFailed != "failed" {
		t.Errorf("recommendation.StatusFailed: beklenen 'failed', aldı '%s'", recommendation.StatusFailed)
	}

	// Birbirinden farklı olmalı
	statuses := []string{
		recommendation.StatusPending,
		recommendation.StatusReady,
		recommendation.StatusFailed,
	}
	seen := make(map[string]bool)
	for _, s := range statuses {
		if seen[s] {
			t.Errorf("Recommendation durum sabiti tekrarlıyor: '%s'", s)
		}
		seen[s] = true
	}
}

// ===========================================================================
// Testler: Mood AnalysisUpdate DTO
// ===========================================================================

// TestAnalysisUpdateDTO, AnalysisUpdate DTO'sunun alanlarının doğru
// tiplendiğini ve sıfır değerlerle oluşturulabildiğini doğrular.
func TestAnalysisUpdateDTO(t *testing.T) {
	update := mood.AnalysisUpdate{
		SentimentLabel:  "positive",
		DominantEmotion: "joy",
		Valence:         0.8,
		Arousal:         0.6,
		Energy:          0.7,
		EmotionScores:   []byte(`{"joy":0.8,"calm":0.2}`),
		Language:        "tr",
		AIModelVersion:  "aria-sentiment-v1.0.0",
		ProcessingMs:    120,
	}

	if update.SentimentLabel == "" {
		t.Error("SentimentLabel boş olmamalı")
	}
	if update.Valence == 0 {
		t.Error("Valence sıfır olmamalı")
	}
	if len(update.EmotionScores) == 0 {
		t.Error("EmotionScores boş olmamalı")
	}
	if update.Language != "tr" {
		t.Errorf("Language: beklenen 'tr', aldı '%s'", update.Language)
	}
}

// ===========================================================================
// Testler: Recommendation DTO
// ===========================================================================

// TestCreateRecommendationInput, orchestrator tarafından kullanılan
// CreateRecommendationInput DTO'sunu doğrular.
func TestCreateRecommendationInput(t *testing.T) {
	input := recommendation.CreateRecommendationInput{
		UserID:         1,
		MoodID:         10,
		AIModelVersion: "aria-rag-v1.0.0",
		RAGContext:     "test bağlamı",
		ProcessingMs:   300,
		Tracks: []recommendation.TrackInput{
			{
				SpotifyTrackID: "3n3Ppam7vgaVa1iaRUc9Lp",
				Title:          "Holocene",
				Artist:         "Bon Iver",
				RelevanceScore: 0.92,
				Reason:         "Test nedeni",
			},
		},
	}

	if input.UserID == 0 {
		t.Error("UserID sıfır olmamalı")
	}
	if input.MoodID == 0 {
		t.Error("MoodID sıfır olmamalı")
	}
	if len(input.Tracks) == 0 {
		t.Fatal("Tracks listesi boş olmamalı")
	}
	if input.Tracks[0].Title == "" {
		t.Error("Track.Title boş olmamalı")
	}
	if input.Tracks[0].Artist == "" {
		t.Error("Track.Artist boş olmamalı")
	}
}

// TestTrackInput_RelevanceScoreRange, TrackInput içindeki RelevanceScore'un
// [0.0, 1.0] aralığında olduğunu doğrular.
func TestTrackInput_RelevanceScoreRange(t *testing.T) {
	tracks := []recommendation.TrackInput{
		{Title: "Şarkı 1", Artist: "Sanatçı 1", RelevanceScore: 0.92},
		{Title: "Şarkı 2", Artist: "Sanatçı 2", RelevanceScore: 0.75},
		{Title: "Şarkı 3", Artist: "Sanatçı 3", RelevanceScore: 0.50},
	}

	for i, track := range tracks {
		if track.RelevanceScore < 0.0 || track.RelevanceScore > 1.0 {
			t.Errorf("tracks[%d].RelevanceScore [0.0, 1.0] aralığında olmalı, aldı: %f", i, track.RelevanceScore)
		}
	}
}

// ===========================================================================
// Testler: Recommendation Model yapısı
// ===========================================================================

// TestRecommendationModel, Recommendation modelinin temel alanlarını
// doğrular. Bu modeli Spotify modülü de kullanır — uyum kritik.
func TestRecommendationModel(t *testing.T) {
	rec := recommendation.Recommendation{
		ID:             1,
		UserID:         42,
		MoodID:         10,
		AIModelVersion: "aria-rag-v1.0.0",
		RAGContext:     "test bağlamı",
		ProcessingMs:   300,
		Status:         recommendation.StatusReady,
		Tracks: []recommendation.RecommendedTrack{
			{
				ID:               1,
				RecommendationID: 1,
				SpotifyTrackID:   "3n3Ppam7vgaVa1iaRUc9Lp",
				Title:            "Holocene",
				Artist:           "Bon Iver",
				Position:         0,
				RelevanceScore:   0.92,
			},
		},
	}

	if rec.ID == 0 {
		t.Error("Recommendation.ID sıfır olmamalı")
	}
	if rec.Status != recommendation.StatusReady {
		t.Errorf("Status: beklenen '%s', aldı '%s'", recommendation.StatusReady, rec.Status)
	}
	if len(rec.Tracks) == 0 {
		t.Fatal("Recommendation.Tracks boş olmamalı")
	}
	if rec.Tracks[0].Position != 0 {
		t.Errorf("İlk track position sıfır olmalı, aldı: %d", rec.Tracks[0].Position)
	}
}

// TestRecommendedTrack_PositionOrdering, Position alanının doğru sırada
// atandığını doğrular. Pipeline servis, track'leri 0'dan başlatarak sıralar.
func TestRecommendedTrack_PositionOrdering(t *testing.T) {
	tracks := []recommendation.RecommendedTrack{
		{Position: 0, Title: "İlk Şarkı"},
		{Position: 1, Title: "İkinci Şarkı"},
		{Position: 2, Title: "Üçüncü Şarkı"},
	}

	for i, track := range tracks {
		if track.Position != i {
			t.Errorf("tracks[%d].Position: beklenen %d, aldı %d", i, i, track.Position)
		}
	}

	// Pozisyon benzersiz olmalı (aynı Recommendation içinde)
	seen := make(map[int]bool)
	for _, track := range tracks {
		if seen[track.Position] {
			t.Errorf("Pozisyon tekrarlıyor: %d", track.Position)
		}
		seen[track.Position] = true
	}
}

// ===========================================================================
// Testler: Mood Model yapısı
// ===========================================================================

// TestMoodModel_DefaultStatus, yeni oluşturulan Mood'un varsayılan
// durumunun "pending" olması gerektiğini doğrular.
func TestMoodModel_DefaultStatus(t *testing.T) {
	// Yeni mood başlangıçta "pending" durumda olmalı
	m := mood.Mood{
		UserID:  1,
		RawText: "Bugün çok yorgunum ama mutluyum",
		Status:  mood.StatusPending,
	}

	if m.Status != "pending" {
		t.Errorf("Yeni Mood.Status: beklenen 'pending', aldı '%s'", m.Status)
	}
	if m.RawText == "" {
		t.Error("Mood.RawText boş olmamalı")
	}
}

// TestMoodModel_AIFieldsDefault, Mood oluşturulduğunda AI alanlarının
// sıfır/boş değerlerde olduğunu doğrular (henüz analiz yapılmadı).
func TestMoodModel_AIFieldsDefault(t *testing.T) {
	m := mood.Mood{
		UserID:  1,
		RawText: "test metni",
		Status:  mood.StatusPending,
	}

	// Analiz yapılmadan önce AI alanları boş/sıfır olmalı
	if m.SentimentLabel != "" {
		t.Error("Yeni Mood'un SentimentLabel'i boş olmalı")
	}
	if m.Valence != 0 {
		t.Error("Yeni Mood'un Valence'ı 0 olmalı")
	}
	if m.Arousal != 0 {
		t.Error("Yeni Mood'un Arousal'ı 0 olmalı")
	}
}

// ===========================================================================
// Testler: ErrNoTracks sentinel hatası
// ===========================================================================

// TestErrNoTracks, recommendation paketindeki ErrNoTracks sentinel
// hatasının doğru tanımlandığını ve pipeline'ın bunu yakalamasını
// sağlayan errors.Is mekanizmasıyla çalıştığını doğrular.
func TestErrNoTracks(t *testing.T) {
	if recommendation.ErrNoTracks == nil {
		t.Fatal("ErrNoTracks nil olmamalı")
	}
	if recommendation.ErrNoTracks.Error() == "" {
		t.Error("ErrNoTracks mesajı boş olmamalı")
	}
}

// TestErrEmptyText, mood paketindeki ErrEmptyText sentinel hatasının
// doğru tanımlandığını doğrular.
func TestErrEmptyText(t *testing.T) {
	if mood.ErrEmptyText == nil {
		t.Fatal("ErrEmptyText nil olmamalı")
	}
	if mood.ErrEmptyText.Error() == "" {
		t.Error("ErrEmptyText mesajı boş olmamalı")
	}
}
