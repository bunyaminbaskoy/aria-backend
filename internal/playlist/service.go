package playlist

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(userID uint, name string, tracks []TrackInput) (*Playlist, error) {
	p := &Playlist{
		UserID: userID,
		Name:   name,
	}
	for i, t := range tracks {
		p.Tracks = append(p.Tracks, PlaylistTrack{
			Position:       i,
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
	if err := s.repo.Create(p); err != nil {
		return nil, err
	}
	return s.repo.FindByID(p.ID)
}

func (s *Service) GetByID(id uint) (*Playlist, error) {
	return s.repo.FindByID(id)
}

func (s *Service) GetUserPlaylists(userID uint) ([]Playlist, error) {
	return s.repo.FindByUserID(userID)
}

func (s *Service) Rename(id uint, userID uint, name string) (*Playlist, error) {
	p, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if p.UserID != userID {
		return nil, ErrNotOwner
	}
	p.Name = name
	if err := s.repo.Update(p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *Service) Delete(id uint, userID uint) error {
	p, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}
	if p.UserID != userID {
		return ErrNotOwner
	}
	return s.repo.Delete(id)
}

func (s *Service) RemoveTrack(playlistID, trackID, userID uint) error {
	p, err := s.repo.FindByID(playlistID)
	if err != nil {
		return err
	}
	if p.UserID != userID {
		return ErrNotOwner
	}
	return s.repo.RemoveTrack(playlistID, trackID)
}

type TrackInput struct {
	SpotifyTrackID string  `json:"spotify_track_id"`
	Title          string  `json:"title"`
	Artist         string  `json:"artist"`
	Album          string  `json:"album"`
	PreviewURL     string  `json:"preview_url"`
	ExternalURL    string  `json:"external_url"`
	DurationMs     int     `json:"duration_ms"`
	RelevanceScore float64 `json:"relevance_score"`
	Reason         string  `json:"reason"`
}
