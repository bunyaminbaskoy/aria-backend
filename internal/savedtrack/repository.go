package savedtrack

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Save(track *SavedTrack) error {
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}, {Name: "spotify_track_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"title", "artist", "album", "preview_url", "external_url", "duration_ms", "mood_key"}),
	}).Create(track).Error
}

func (r *Repository) FindByUserID(userID uint) ([]SavedTrack, error) {
	var tracks []SavedTrack
	err := r.db.Where("user_id = ?", userID).Order("created_at DESC").Find(&tracks).Error
	return tracks, err
}

func (r *Repository) FindByUserIDAndMood(userID uint, moodKey string) ([]SavedTrack, error) {
	var tracks []SavedTrack
	err := r.db.Where("user_id = ? AND mood_key = ?", userID, moodKey).Order("created_at DESC").Find(&tracks).Error
	return tracks, err
}

func (r *Repository) DeleteByUserAndTrack(userID uint, spotifyTrackID string) error {
	return r.db.Where("user_id = ? AND spotify_track_id = ?", userID, spotifyTrackID).Delete(&SavedTrack{}).Error
}

func (r *Repository) CountByUserID(userID uint) (int64, error) {
	var count int64
	err := r.db.Model(&SavedTrack{}).Where("user_id = ?", userID).Count(&count).Error
	return count, err
}
