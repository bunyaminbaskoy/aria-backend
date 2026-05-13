package pipeline

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"

	"music-curation/internal/mood"
	"music-curation/internal/recommendation"
	"music-curation/pkg/aiclient"
)

// AIClient, pipeline'ın AI servisinden ihtiyaç duyduğu davranışı
// tanımlayan minimal arayüzdür. *aiclient.Client somut tipi yerine
// arayüze bağımlı olmak:
//
//   - Test sırasında mock AI implementasyonu enjekte etmeyi mümkün kılar.
//   - aiclient paketinin iç değişikliklerinden pipeline'ı izole eder.
//   - Modular monolith disiplinine uygundur (modüller davranışa
//     bağımlı, somut tipe değil).
type AIClient interface {
	AnalyzeMood(ctx context.Context, req aiclient.AnalyzeRequest) (*aiclient.AnalyzeResponse, error)
	GetRecommendations(ctx context.Context, req aiclient.RecommendRequest) (*aiclient.RecommendResponse, error)
}

// Pipeline-level sentinel error'lar. Handler bunları HTTP status'larına
// eşler. errors.Is ile alt katmandaki hata kategorileri (örn. aiclient
// sentinel'ları) buraya kadar zincirli şekilde taşınır.
var (
	// ErrEmptyText, kullanıcı boş metin gönderdiğinde döner.
	ErrEmptyText = errors.New("ruh hali metni boş olamaz")

	// ErrAIUnavailable, AI servisi erişilemez durumdayken döner.
	// aiclient.ErrServiceUnavailable bunun altına maplenir.
	ErrAIUnavailable = errors.New("AI servisi şu anda kullanılamıyor")

	// ErrAIBadRequest, AI servisi 4xx döndürdüğünde döner.
	ErrAIBadRequest = errors.New("AI servisi isteği reddetti")

	// ErrAIInternal, AI servisi 5xx döndürdüğünde döner.
	ErrAIInternal = errors.New("AI servisinde dahili hata")

	// ErrPersistence, veritabanına yazma sırasında oluşan hatalar için.
	ErrPersistence = errors.New("kayıt veritabanına yazılamadı")
)

// defaultRecommendationLimit, request'te limit verilmediğinde kullanılır.
const defaultRecommendationLimit = 20

// moodPresets, mood_key ile gelen isteklerde /analyze çağrısını atlayarak
// doğrudan kullanılacak önceden tanımlı analiz sonuçlarıdır.
var moodPresets = map[string]*aiclient.AnalyzeResponse{
	"happy": {
		SentimentLabel: "positive", DominantEmotion: "joy",
		Valence: 0.8, Arousal: 0.5, Energy: 0.7,
		EmotionScores: map[string]float64{"joy": 0.9, "excitement": 0.6},
		Language: "tr", ModelVersion: "preset-v1.0.0",
	},
	"sad": {
		SentimentLabel: "negative", DominantEmotion: "sadness",
		Valence: -0.6, Arousal: -0.4, Energy: 0.2,
		EmotionScores: map[string]float64{"sadness": 0.8, "melancholy": 0.7},
		Language: "tr", ModelVersion: "preset-v1.0.0",
	},
	"angry": {
		SentimentLabel: "negative", DominantEmotion: "anger",
		Valence: -0.5, Arousal: 0.8, Energy: 0.9,
		EmotionScores: map[string]float64{"anger": 0.9, "frustration": 0.7},
		Language: "tr", ModelVersion: "preset-v1.0.0",
	},
	"relaxed": {
		SentimentLabel: "positive", DominantEmotion: "calm",
		Valence: 0.4, Arousal: -0.6, Energy: 0.2,
		EmotionScores: map[string]float64{"calm": 0.9, "contentment": 0.6},
		Language: "tr", ModelVersion: "preset-v1.0.0",
	},
	"energetic": {
		SentimentLabel: "positive", DominantEmotion: "excitement",
		Valence: 0.7, Arousal: 0.9, Energy: 0.95,
		EmotionScores: map[string]float64{"excitement": 0.9, "joy": 0.5},
		Language: "tr", ModelVersion: "preset-v1.0.0",
	},
	"romantic": {
		SentimentLabel: "positive", DominantEmotion: "love",
		Valence: 0.6, Arousal: 0.1, Energy: 0.4,
		EmotionScores: map[string]float64{"love": 0.9, "tenderness": 0.7},
		Language: "tr", ModelVersion: "preset-v1.0.0",
	},
	"nostalgic": {
		SentimentLabel: "mixed", DominantEmotion: "nostalgia",
		Valence: 0.1, Arousal: -0.2, Energy: 0.3,
		EmotionScores: map[string]float64{"nostalgia": 0.9, "melancholy": 0.4},
		Language: "tr", ModelVersion: "preset-v1.0.0",
	},
	"focused": {
		SentimentLabel: "neutral", DominantEmotion: "concentration",
		Valence: 0.2, Arousal: 0.3, Energy: 0.5,
		EmotionScores: map[string]float64{"focus": 0.9, "calm": 0.5},
		Language: "tr", ModelVersion: "preset-v1.0.0",
	},
}

// Service, sentiment analizi + RAG öneri üretimi pipeline'ının ana
// orchestrator'ıdır. Modüller arası iletişimin tek bir noktada
// toplandığı yerdir; tüm bağımlılıklar Go fonksiyon çağrıları ile
// (ağ üzerinden DEĞİL) yapılır.
type Service struct {
	moodService *mood.Service
	recService  *recommendation.Service
	aiClient    AIClient
}

// NewService, tüm bağımlılıkları enjekte ederek yeni bir orchestrator
// üretir. main.go içinde tek seferlik çağrılır.
func NewService(
	moodService *mood.Service,
	recService *recommendation.Service,
	aiClient AIClient,
) *Service {
	return &Service{
		moodService: moodService,
		recService:  recService,
		aiClient:    aiClient,
	}
}

// GeneratePlaylist, sistemin ana giriş noktası. Adımlar:
//
//	a) Ham metni veritabanına "pending" Mood olarak kaydet.
//	b) AI'a sentiment analizi yaptır.
//	c) Mood kaydını analiz sonuçlarıyla güncelle.
//	d) AI'dan parça önerileri al.
//	e) Recommendation + tracks kayıtlarını yaz.
//	f) Birleştirilmiş sonucu döndür.
//
// Hata davranışı:
//   - (a) başarısızsa: hiç kayıt kalmaz, hata döner.
//   - (b) başarısızsa: Mood "failed" olarak işaretlenir, hata döner.
//   - (c) başarısızsa: Mood "pending"de kalır (idempotent re-run mümkün).
//   - (d) başarısızsa: Mood "analyzed"de kalır, recommendation yazılmaz.
//   - (e) başarısızsa: Mood "analyzed"de kalır.
//
// Bu adımcı yaklaşım, kısmi başarıları gözlemlenebilir kılar; aynı
// metin için "tekrar dene" mantığı kolayca eklenebilir.
func (s *Service) GeneratePlaylist(ctx context.Context, userID uint, rawText string, moodKey string, limit int) (*GenerateResponse, error) {
	// Erken doğrulama — en az bir giriş gerekli.
	trimmed := strings.TrimSpace(rawText)
	moodKey = strings.TrimSpace(strings.ToLower(moodKey))

	if trimmed == "" && moodKey == "" {
		return nil, ErrEmptyText
	}
	if limit <= 0 {
		limit = defaultRecommendationLimit
	}

	// mood_key varsa ve text yoksa, rawText olarak mood_key kullan.
	if trimmed == "" {
		trimmed = moodKey
	}

	// (a) Pending Mood kaydı oluştur.
	m, err := s.moodService.CreateRawMood(userID, trimmed)
	if err != nil {
		if errors.Is(err, mood.ErrEmptyText) {
			return nil, ErrEmptyText
		}
		return nil, fmt.Errorf("%w: mood: %v", ErrPersistence, err)
	}

	// (b) Sentiment analizi — mood_key varsa preset kullan, yoksa AI'a sor.
	var analysis *aiclient.AnalyzeResponse

	if moodKey != "" {
		preset, ok := moodPresets[moodKey]
		if !ok {
			return nil, fmt.Errorf("%w: geçersiz mood_key: %s", ErrEmptyText, moodKey)
		}
		analysis = preset
	} else {
		analyzeReq := aiclient.AnalyzeRequest{
			Text:   m.RawText,
			UserID: userID,
		}
		analysis, err = s.aiClient.AnalyzeMood(ctx, analyzeReq)
		if err != nil {
			if mfErr := s.moodService.MarkFailed(m.ID); mfErr != nil {
				log.Printf("⚠️  Mood %d 'failed' olarak işaretlenemedi: %v", m.ID, mfErr)
			}
			return nil, mapAIError(err, "analyze")
		}
	}

	// (c) Mood kaydını analiz sonuçlarıyla güncelle.
	emotionScores, _ := json.Marshal(analysis.EmotionScores)
	if err := s.moodService.UpdateAnalysis(m.ID, mood.AnalysisUpdate{
		SentimentLabel:  analysis.SentimentLabel,
		DominantEmotion: analysis.DominantEmotion,
		Valence:         analysis.Valence,
		Arousal:         analysis.Arousal,
		Energy:          analysis.Energy,
		EmotionScores:   emotionScores,
		Language:        analysis.Language,
		AIModelVersion:  analysis.ModelVersion,
		ProcessingMs:    analysis.ProcessingMs,
	}); err != nil {
		return nil, fmt.Errorf("%w: mood analiz güncellemesi: %v", ErrPersistence, err)
	}

	// Mood'u taze haliyle yeniden yükle — response'ta güncel alanlarla dönsün.
	m, err = s.moodService.GetMoodByID(m.ID)
	if err != nil || m == nil {
		return nil, fmt.Errorf("%w: mood reload: %v", ErrPersistence, err)
	}

	// (d) AI'dan parça önerisi iste.
	recReq := aiclient.RecommendRequest{
		UserID: userID,
		MoodID: m.ID,
		Mood: aiclient.MoodSnapshot{
			SentimentLabel:  analysis.SentimentLabel,
			DominantEmotion: analysis.DominantEmotion,
			Valence:         analysis.Valence,
			Arousal:         analysis.Arousal,
			Energy:          analysis.Energy,
		},
		Limit: limit,
		// Context şimdilik nil — gelecekteki sprint'te kullanıcı tercihleri
		// (geçmiş dinleme, exclude listesi) burada doldurulacak.
	}
	recResp, err := s.aiClient.GetRecommendations(ctx, recReq)
	if err != nil {
		return nil, mapAIError(err, "recommend")
	}

	// (e) Recommendation + tracks'i tek transaction'da yaz.
	tracks := make([]recommendation.TrackInput, 0, len(recResp.Tracks))
	for _, t := range recResp.Tracks {
		tracks = append(tracks, recommendation.TrackInput{
			SpotifyTrackID: t.SpotifyTrackID,
			Title:          t.Title,
			Artist:         t.Artist,
			Album:          t.Album,
			PreviewURL:     t.PreviewURL,
			ExternalURL:    t.ExternalURL,
			DurationMs:     t.DurationMs,
			RelevanceScore: t.RelevanceScore,
			Reason:         t.Reason,
		})
	}
	rec, err := s.recService.CreateFromAI(recommendation.CreateRecommendationInput{
		UserID:         userID,
		MoodID:         m.ID,
		AIModelVersion: recResp.ModelVersion,
		RAGContext:     recResp.RAGContext,
		ProcessingMs:   recResp.ProcessingMs,
		Tracks:         tracks,
	})
	if err != nil {
		if errors.Is(err, recommendation.ErrNoTracks) {
			// AI boş öneri döndü — bu pipeline için anlamlı bir hatadır.
			return nil, fmt.Errorf("%w: AI boş öneri listesi döndürdü", ErrAIInternal)
		}
		return nil, fmt.Errorf("%w: recommendation: %v", ErrPersistence, err)
	}

	// (f) Birleştirilmiş sonucu döndür.
	return &GenerateResponse{
		Mood:           m,
		Recommendation: rec,
	}, nil
}

// mapAIError, aiclient katmanından gelen sentinel hatalarını
// pipeline-level sentinel'larına çevirir. Stage parametresi
// (analyze/recommend) hata zincirinde hangi adımda olduğumuzu açıklar.
func mapAIError(err error, stage string) error {
	switch {
	case errors.Is(err, aiclient.ErrServiceUnavailable):
		return fmt.Errorf("%w (%s): %v", ErrAIUnavailable, stage, err)
	case errors.Is(err, aiclient.ErrInvalidRequest):
		return fmt.Errorf("%w (%s): %v", ErrAIBadRequest, stage, err)
	case errors.Is(err, aiclient.ErrInternal):
		return fmt.Errorf("%w (%s): %v", ErrAIInternal, stage, err)
	case errors.Is(err, aiclient.ErrDecode):
		return fmt.Errorf("%w (%s): %v", ErrAIInternal, stage, err)
	default:
		return fmt.Errorf("%w (%s): %v", ErrAIInternal, stage, err)
	}
}
