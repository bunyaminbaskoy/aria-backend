package recommendation

import (
	"time"
)

// Recommendation, bir Mood kaydı için Python AI servisinden dönen
// öneri kümesinin "başlık" (header) kaydıdır. Her öneri kümesi
// 1..N adet RecommendedTrack içerir.
//
// Bir Mood için birden fazla Recommendation oluşabilir (örn. kullanıcı
// "yeniden öner" istediğinde), bu yüzden ilişki 1:N olarak modellendi.
//
// AI servisinin model versiyonunu ve RAG context'ini saklamak,
// gelecekte aynı sonuçları reproduce edebilmek ve A/B karşılaştırma
// yapabilmek için kritik öneme sahiptir.
type Recommendation struct {
	ID     uint `gorm:"primaryKey" json:"id"`
	UserID uint `gorm:"not null;index" json:"user_id"`
	MoodID uint `gorm:"not null;index" json:"mood_id"`

	// AIModelVersion, öneriyi üreten RAG modelinin sürümüdür.
	// Örn: "aria-rag-v1.2.0".
	AIModelVersion string `gorm:"size:64" json:"ai_model_version,omitempty"`

	// RAGContext, AI servisinin öneriyi üretirken kullandığı bağlam
	// metnidir (LLM prompt'u veya retrieve edilen dokümanların özeti).
	// Açıklanabilirlik (explainability) ve debug için saklanır.
	RAGContext string `gorm:"type:text" json:"rag_context,omitempty"`

	// ProcessingMs, /recommend çağrısının toplam süresi (milisaniye).
	ProcessingMs int `gorm:"default:0" json:"processing_ms"`

	// Status, öneri kümesinin durumunu izler. Asenkron pipeline'a
	// geçildiğinde "pending" -> "ready" / "failed" akışı kullanılır.
	//   "pending" → AI çağrısı başlatıldı, sonuç bekleniyor
	//   "ready"   → tracks başarıyla yazıldı
	//   "failed"  → AI çağrısı başarısız oldu
	Status string `gorm:"size:16;not null;default:'pending';index" json:"status"`

	// Tracks, has-many ilişkisidir. Recommendation silindiğinde alt
	// parçaların da silinmesi için OnDelete:CASCADE tanımlandı.
	// Preload ile birlikte yüklenir; aksi halde JSON'da boş gelir.
	Tracks []RecommendedTrack `gorm:"foreignKey:RecommendationID;constraint:OnDelete:CASCADE" json:"tracks,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// RecommendedTrack, bir Recommendation içindeki tekil bir parça önerisidir.
//
// SpotifyTrackID şu anda nullable tutuldu: AI servisi başlangıçta
// sadece başlık/sanatçı önerebilir, Spotify ID'si daha sonra
// "Spotify search" adımında doldurulabilir. Bu sayede iki ekip
// (AI ekibi ve Spotify ekibi) birbirinden bağımsız ilerleyebilir.
type RecommendedTrack struct {
	ID uint `gorm:"primaryKey" json:"id"`

	// RecommendationID, parent Recommendation'ın ID'si.
	// Composite unique index'in ilk kolonudur — (recommendation_id, position)
	// çifti benzersiz olur, böylece aynı küme içinde Position çakışması olmaz.
	RecommendationID uint `gorm:"not null;index;uniqueIndex:idx_reco_position" json:"recommendation_id"`

	// SpotifyTrackID, Spotify'ın katalog kimliğidir (örn. "3n3Ppam7vgaVa1iaRUc9Lp").
	// Henüz çözümlenmediyse boş bırakılır.
	SpotifyTrackID string `gorm:"size:64;index" json:"spotify_track_id,omitempty"`

	Title  string `gorm:"size:255;not null" json:"title"`
	Artist string `gorm:"size:255;not null" json:"artist"`
	Album  string `gorm:"size:255" json:"album,omitempty"`

	PreviewURL  string `gorm:"size:512" json:"preview_url,omitempty"`
	ExternalURL string `gorm:"size:512" json:"external_url,omitempty"`

	DurationMs int `gorm:"default:0" json:"duration_ms"`

	// Position, öneri kümesi içindeki sıradır (0'dan başlar).
	// Composite unique index'in ikinci kolonu — bkz. RecommendationID.
	Position int `gorm:"not null;uniqueIndex:idx_reco_position" json:"position"`

	// RelevanceScore, AI'nın bu parçanın ruh hali ile ne kadar
	// örtüştüğüne dair skoru (0.0 - 1.0). Sıralama ve filtreleme için.
	RelevanceScore float64 `gorm:"default:0" json:"relevance_score"`

	// Reason, AI'nın bu parçanın neden önerildiğine dair ürettiği
	// kısa açıklamadır. Frontend'de "neden bu şarkı?" ipucu için.
	Reason string `gorm:"type:text" json:"reason,omitempty"`

	CreatedAt time.Time `json:"created_at"`
}

// Recommendation durum sabitleri — sihirli string kullanmamak için.
const (
	StatusPending = "pending"
	StatusReady   = "ready"
	StatusFailed  = "failed"
)
