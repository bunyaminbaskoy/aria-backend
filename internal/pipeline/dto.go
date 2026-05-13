package pipeline

import (
	"music-curation/internal/mood"
	"music-curation/internal/recommendation"
)

// GenerateRequest, POST /api/v1/recommendations/generate endpoint'ine
// gelen request gövdesini temsil eder. Kullanıcı kimliği body'de
// taşınmaz — AuthMiddleware tarafından context'e konulan userID kullanılır.
//
// Örnek istek:
//
//	{
//	  "text": "Bugün bahar gibiyim, biraz dansa ihtiyacım var"
//	}
type GenerateRequest struct {
	// Text, kullanıcının serbest metin girişi. MoodKey boşsa zorunludur.
	Text string `json:"text" binding:"omitempty,max=2000"`

	// MoodKey, önceden tanımlı 8 ruh halinden biri: happy, sad, angry,
	// relaxed, energetic, romantic, nostalgic, focused.
	// Text boşsa zorunludur.
	MoodKey string `json:"mood_key" binding:"omitempty"`

	// Limit, opsiyoneldir. AI servisinden istenecek parça sayısı.
	// Verilmezse veya 0 ise pipeline varsayılanı (20) kullanılır.
	Limit int `json:"limit,omitempty" binding:"omitempty,min=1,max=50"`
}

// GenerateResponse, orchestrator'ın frontend'e döndüğü tek atomik
// yanıt yapısıdır. İçinde hem analiz edilmiş Mood, hem de üretilen
// Recommendation (parçalarıyla birlikte) bulunur.
//
// Bu yapı sayesinde frontend tek bir HTTP çağrısıyla tüm flow'un
// sonucunu alır; ayrıca GET /moods/:id ve GET /recommendations/:id
// gibi takip çağrıları yapmaya gerek kalmaz.
type GenerateResponse struct {
	Mood           *mood.Mood                     `json:"mood"`
	Recommendation *recommendation.Recommendation `json:"recommendation"`
}
