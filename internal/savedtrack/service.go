package savedtrack

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Save(userID uint, req SaveRequest) (*SavedTrack, error) {
	track := &SavedTrack{
		UserID:         userID,
		SpotifyTrackID: req.SpotifyTrackID,
		Title:          req.Title,
		Artist:         req.Artist,
		Album:          req.Album,
		PreviewURL:     req.PreviewURL,
		ExternalURL:    req.ExternalURL,
		DurationMs:     req.DurationMs,
		MoodKey:        req.MoodKey,
	}
	if err := s.repo.Save(track); err != nil {
		return nil, err
	}
	return track, nil
}

func (s *Service) GetUserSaved(userID uint) ([]SavedTrack, error) {
	return s.repo.FindByUserID(userID)
}

func (s *Service) GetUserSavedByMood(userID uint, moodKey string) ([]SavedTrack, error) {
	return s.repo.FindByUserIDAndMood(userID, moodKey)
}

func (s *Service) Unsave(userID uint, spotifyTrackID string) error {
	return s.repo.DeleteByUserAndTrack(userID, spotifyTrackID)
}

func (s *Service) Count(userID uint) (int64, error) {
	return s.repo.CountByUserID(userID)
}

type SaveRequest struct {
	SpotifyTrackID string `json:"spotify_track_id" binding:"required"`
	Title          string `json:"title" binding:"required"`
	Artist         string `json:"artist" binding:"required"`
	Album          string `json:"album"`
	PreviewURL     string `json:"preview_url"`
	ExternalURL    string `json:"external_url"`
	DurationMs     int    `json:"duration_ms"`
	MoodKey        string `json:"mood_key"`
}
