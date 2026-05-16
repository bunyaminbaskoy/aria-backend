package playlist

import "gorm.io/gorm"

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(p *Playlist) error {
	return r.db.Create(p).Error
}

func (r *Repository) FindByID(id uint) (*Playlist, error) {
	var p Playlist
	err := r.db.Preload("Tracks", func(db *gorm.DB) *gorm.DB {
		return db.Order("position ASC")
	}).First(&p, id).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *Repository) FindByUserID(userID uint) ([]Playlist, error) {
	var playlists []Playlist
	err := r.db.Preload("Tracks", func(db *gorm.DB) *gorm.DB {
		return db.Order("position ASC")
	}).Where("user_id = ?", userID).Order("created_at DESC").Find(&playlists).Error
	return playlists, err
}

func (r *Repository) Update(p *Playlist) error {
	return r.db.Save(p).Error
}

func (r *Repository) Delete(id uint) error {
	return r.db.Delete(&Playlist{}, id).Error
}

func (r *Repository) AddTrack(t *PlaylistTrack) error {
	return r.db.Create(t).Error
}

func (r *Repository) RemoveTrack(playlistID, trackID uint) error {
	return r.db.Where("playlist_id = ? AND id = ?", playlistID, trackID).Delete(&PlaylistTrack{}).Error
}

func (r *Repository) GetMaxPosition(playlistID uint) (int, error) {
	var max int
	err := r.db.Model(&PlaylistTrack{}).
		Where("playlist_id = ?", playlistID).
		Select("COALESCE(MAX(position), -1)").
		Scan(&max).Error
	return max, err
}
