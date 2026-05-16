package interaction

import (
	"time"
)

// TrackInteraction, bir kullanıcının önerilen bir parçayla etkileşimini
// temsil eder (beğenme veya beğenmeme). Collaborative filtering için
// temel veri kaynağıdır.
//
// Composite unique index (user_id, spotify_track_id) sayesinde bir
// kullanıcı aynı parça için yalnızca tek bir etkileşim kaydı tutar;
// yeni etkileşim geldiğinde UPSERT ile güncellenir.
type TrackInteraction struct {
	ID uint `gorm:"primaryKey" json:"id"`

	// UserID, etkileşimi yapan kullanıcının iç ID'si.
	UserID uint `gorm:"not null;index;uniqueIndex:idx_user_track" json:"user_id"`

	// SpotifyTrackID, etkileşim yapılan parçanın Spotify katalog ID'si.
	SpotifyTrackID string `gorm:"size:64;not null;uniqueIndex:idx_user_track" json:"spotify_track_id"`

	// RecommendationID, bu etkileşimi tetikleyen öneri kümesinin ID'si.
	// Opsiyoneldir; analitik ve debug için saklanır.
	RecommendationID uint `gorm:"index" json:"recommendation_id,omitempty"`

	// InteractionType, etkileşim türüdür: "like" veya "dislike".
	InteractionType string `gorm:"size:16;not null;index" json:"interaction_type"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Etkileşim türü sabitleri.
const (
	TypeLike    = "like"
	TypeDislike = "dislike"
)

// CoLikedResult, collaborative filtering co-occurrence sorgusunun
// döndürdüğü tek bir satırı temsil eder.
type CoLikedResult struct {
	SpotifyTrackID string `json:"spotify_track_id"`
	CoLikeCount    int    `json:"co_like_count"`
}
