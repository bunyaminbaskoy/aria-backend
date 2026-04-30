package mood

import (
	"time"

	"gorm.io/datatypes"
)

// Mood, kullanıcının girdiği ham metni ve Python AI servisinin
// bu metin için ürettiği duygu/sentiment analizi sonuçlarını saklar.
//
// Yaşam döngüsü:
//  1. Kullanıcı POST /moods ile ham metin gönderir.
//  2. Önce yalnızca RawText doldurulmuş bir kayıt yaratılır (Status = "pending").
//  3. Orchestrator pipeline'ı (sonraki adımda eklenecek) Python AI servisini
//     çağırır ve sentiment alanlarını (Valence, Arousal, EmotionScores, vb.)
//     bu kayda yazar (Status = "analyzed").
//  4. Aynı Mood kaydı, Recommendation modülü tarafından öneri üretmek için
//     girdi olarak kullanılır.
//
// Bu yapı, AI çağrısı senkron veya asenkron yapılsa da aynı şekilde çalışır.
type Mood struct {
	ID     uint `gorm:"primaryKey" json:"id"`
	UserID uint `gorm:"not null;index" json:"user_id"`

	RawText string `gorm:"type:text;not null" json:"raw_text"`

	// AI tarafından üretilen alanlar — analiz tamamlanana kadar boş/sıfırdır.
	SentimentLabel  string  `gorm:"size:64" json:"sentiment_label,omitempty"`
	DominantEmotion string  `gorm:"size:64" json:"dominant_emotion,omitempty"`
	Valence         float64 `gorm:"default:0" json:"valence"`
	Arousal         float64 `gorm:"default:0" json:"arousal"`
	Energy          float64 `gorm:"default:0" json:"energy"`

	// EmotionScores, AI servisinden gelen tüm duygu skorlarını JSONB olarak tutar.
	// Örnek: {"joy": 0.6, "sadness": 0.2, "fatigue": 0.7}
	// Esnek bir alan olduğu için ilişkisel kolon yerine JSONB tercih edildi.
	EmotionScores datatypes.JSON `gorm:"type:jsonb" json:"emotion_scores,omitempty"`

	// Language, AI'ın tespit ettiği dil kodu (ISO 639-1). Örn: "tr", "en".
	Language string `gorm:"size:8" json:"language,omitempty"`

	// AIModelVersion, hangi model sürümünün bu analizi ürettiğini izlemek
	// için kullanılır. Reprocessing kararları (eski sürümleri yeniden işleme)
	// bu alana bakılarak verilir.
	AIModelVersion string `gorm:"size:64" json:"ai_model_version,omitempty"`

	// ProcessingMs, AI çağrısının ne kadar sürdüğünü milisaniye cinsinden tutar.
	// Performans izleme ve debug için faydalıdır.
	ProcessingMs int `gorm:"default:0" json:"processing_ms"`

	// Status, kaydın yaşam döngüsündeki konumunu belirtir:
	//   "pending"  → ham metin alındı, AI çağrısı henüz yapılmadı
	//   "analyzed" → AI sonuçları başarıyla yazıldı
	//   "failed"   → AI çağrısı başarısız oldu (RawText korunur)
	Status string `gorm:"size:16;not null;default:'pending';index" json:"status"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Mood durumları için sabitler — sihirli string kullanmamak için.
const (
	StatusPending  = "pending"
	StatusAnalyzed = "analyzed"
	StatusFailed   = "failed"
)
