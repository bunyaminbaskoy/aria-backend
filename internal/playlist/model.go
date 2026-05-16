package playlist

import "time"

type Playlist struct {
	ID     uint   `gorm:"primaryKey" json:"id"`
	UserID uint   `gorm:"not null;index" json:"user_id"`
	Name   string `gorm:"size:255;not null" json:"name"`

	Tracks []PlaylistTrack `gorm:"foreignKey:PlaylistID;constraint:OnDelete:CASCADE" json:"tracks,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type PlaylistTrack struct {
	ID         uint   `gorm:"primaryKey" json:"id"`
	PlaylistID uint   `gorm:"not null;index;uniqueIndex:idx_playlist_position" json:"playlist_id"`
	Position   int    `gorm:"not null;uniqueIndex:idx_playlist_position" json:"position"`

	SpotifyTrackID string `gorm:"size:64;index" json:"spotify_track_id,omitempty"`
	Title          string `gorm:"size:255;not null" json:"title"`
	Artist         string `gorm:"size:255;not null" json:"artist"`
	Album          string `gorm:"size:255" json:"album,omitempty"`
	PreviewURL     string `gorm:"size:512" json:"preview_url,omitempty"`
	ExternalURL    string `gorm:"size:512" json:"external_url,omitempty"`
	DurationMs     int    `gorm:"default:0" json:"duration_ms"`
	RelevanceScore float64 `gorm:"default:0" json:"relevance_score"`
	Reason         string `gorm:"type:text" json:"reason,omitempty"`

	CreatedAt time.Time `json:"created_at"`
}
