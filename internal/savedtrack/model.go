package savedtrack

import "time"

// SavedTrack, kullanıcının kaydettiği (bookmark) bir şarkıyı temsil eder.
type SavedTrack struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	UserID         uint      `gorm:"not null;index;uniqueIndex:idx_user_saved_track" json:"user_id"`
	SpotifyTrackID string    `gorm:"size:64;not null;uniqueIndex:idx_user_saved_track" json:"spotify_track_id"`
	Title          string    `gorm:"size:255;not null" json:"title"`
	Artist         string    `gorm:"size:255;not null" json:"artist"`
	Album          string    `gorm:"size:255" json:"album"`
	PreviewURL     string    `gorm:"size:512" json:"preview_url"`
	ExternalURL    string    `gorm:"size:512" json:"external_url"`
	DurationMs     int       `json:"duration_ms"`
	MoodKey        string    `gorm:"size:32;index" json:"mood_key"`
	CreatedAt      time.Time `json:"created_at"`
}
