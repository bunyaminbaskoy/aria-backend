package interaction

// CreateInteractionRequest, POST /api/v1/interactions endpoint'inin
// istek gövdesidir.
type CreateInteractionRequest struct {
	// SpotifyTrackID, etkileşim yapılan parçanın Spotify ID'si.
	SpotifyTrackID string `json:"spotify_track_id" binding:"required"`

	// RecommendationID, bu etkileşimi tetikleyen öneri kümesinin ID'si.
	// Opsiyoneldir.
	RecommendationID uint `json:"recommendation_id"`

	// InteractionType, etkileşim türü: "like" veya "dislike".
	InteractionType string `json:"interaction_type" binding:"required,oneof=like dislike"`
}
