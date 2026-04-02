package user

import (
	"time"
)

// User represents the users table in PostgreSQL.
// Supports both local (email+password) and OAuth (Google, Spotify) authentication.
type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Email     string    `gorm:"uniqueIndex;not null;size:255" json:"email"`
	Password  string    `gorm:"size:255" json:"-"`                        // Hidden from JSON; optional for OAuth users
	GoogleID  *string   `gorm:"uniqueIndex;size:255" json:"google_id,omitempty"`  // Nullable — for Google OAuth
	SpotifyID           *string `gorm:"uniqueIndex;size:255" json:"spotify_id,omitempty"` // Nullable — for Spotify OAuth
	SpotifyAccessToken  *string `gorm:"size:512" json:"-"`                               // Hidden — for Spotify API calls
	SpotifyRefreshToken *string `gorm:"size:512" json:"-"`                               // Hidden — for token refresh
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
