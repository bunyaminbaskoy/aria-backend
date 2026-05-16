package interaction

import (
	"errors"
)

// ErrInvalidType, geçersiz etkileşim türü gönderildiğinde döner.
var ErrInvalidType = errors.New("etkileşim türü 'like' veya 'dislike' olmalıdır")

// Service, TrackInteraction modülünün iş kuralı katmanıdır.
type Service struct {
	repo Repository
}

// NewService, Repository'i alarak yeni bir interaction service üretir.
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// Upsert, kullanıcının bir parçayla etkileşimini kaydeder veya günceller.
func (s *Service) Upsert(userID uint, req CreateInteractionRequest) (*TrackInteraction, error) {
	if req.InteractionType != TypeLike && req.InteractionType != TypeDislike {
		return nil, ErrInvalidType
	}

	interaction := &TrackInteraction{
		UserID:           userID,
		SpotifyTrackID:   req.SpotifyTrackID,
		RecommendationID: req.RecommendationID,
		InteractionType:  req.InteractionType,
	}

	if err := s.repo.Upsert(interaction); err != nil {
		return nil, err
	}
	return interaction, nil
}

// GetUserInteractions, kullanıcının tüm etkileşimlerini döner.
func (s *Service) GetUserInteractions(userID uint, limit, offset int) ([]TrackInteraction, error) {
	return s.repo.FindByUserID(userID, limit, offset)
}

// DeleteInteraction, bir kullanıcı-parça etkileşimini siler.
func (s *Service) DeleteInteraction(userID uint, spotifyTrackID string) error {
	return s.repo.DeleteByUserAndTrack(userID, spotifyTrackID)
}

// GetLikedTrackIDs, kullanıcının beğendiği parçaların ID'lerini döner.
func (s *Service) GetLikedTrackIDs(userID uint) ([]string, error) {
	return s.repo.GetLikedTrackIDs(userID)
}

// GetDislikedTrackIDs, kullanıcının beğenmediği parçaların ID'lerini döner.
func (s *Service) GetDislikedTrackIDs(userID uint) ([]string, error) {
	return s.repo.GetDislikedTrackIDs(userID)
}

// GetCollabTrackIDs, collaborative filtering ile önerilen parça ID'lerini döner.
func (s *Service) GetCollabTrackIDs(userID uint, limit int) ([]string, error) {
	likedIDs, err := s.repo.GetLikedTrackIDs(userID)
	if err != nil {
		return nil, err
	}
	if len(likedIDs) == 0 {
		return nil, nil
	}

	results, err := s.repo.GetCoLikedTracks(likedIDs, userID, limit)
	if err != nil {
		return nil, err
	}

	ids := make([]string, 0, len(results))
	for _, r := range results {
		ids = append(ids, r.SpotifyTrackID)
	}
	return ids, nil
}
