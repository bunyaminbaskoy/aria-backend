package interaction

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Repository, TrackInteraction verilerine erişim arayüzüdür.
type Repository interface {
	Upsert(interaction *TrackInteraction) error
	FindByUserID(userID uint, limit, offset int) ([]TrackInteraction, error)
	GetLikedTrackIDs(userID uint) ([]string, error)
	GetDislikedTrackIDs(userID uint) ([]string, error)
	DeleteByUserAndTrack(userID uint, spotifyTrackID string) error
	GetCoLikedTracks(likedIDs []string, excludeUserID uint, limit int) ([]CoLikedResult, error)
}

type repository struct {
	db *gorm.DB
}

// NewRepository, GORM bağlantısını alarak yeni bir interaction
// repository üretir.
func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

// Upsert, aynı kullanıcı-parça çifti için mevcut kaydı günceller veya
// yeni kayıt oluşturur. Kullanıcı bir parçayı önce like edip sonra
// dislike ettiğinde, tek bir satırda güncellenir.
func (r *repository) Upsert(interaction *TrackInteraction) error {
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}, {Name: "spotify_track_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"interaction_type", "recommendation_id", "updated_at"}),
	}).Create(interaction).Error
}

// FindByUserID, bir kullanıcının tüm etkileşimlerini en yeniden eskiye
// sıralı olarak döner.
func (r *repository) FindByUserID(userID uint, limit, offset int) ([]TrackInteraction, error) {
	if limit <= 0 {
		limit = 50
	}
	var interactions []TrackInteraction
	result := r.db.
		Where("user_id = ?", userID).
		Order("updated_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&interactions)
	return interactions, result.Error
}

// GetLikedTrackIDs, kullanıcının beğendiği parçaların Spotify ID'lerini döner.
func (r *repository) GetLikedTrackIDs(userID uint) ([]string, error) {
	return r.getTrackIDsByType(userID, TypeLike)
}

// GetDislikedTrackIDs, kullanıcının beğenmediği parçaların Spotify ID'lerini döner.
func (r *repository) GetDislikedTrackIDs(userID uint) ([]string, error) {
	return r.getTrackIDsByType(userID, TypeDislike)
}

func (r *repository) getTrackIDsByType(userID uint, interactionType string) ([]string, error) {
	var ids []string
	result := r.db.Model(&TrackInteraction{}).
		Where("user_id = ? AND interaction_type = ?", userID, interactionType).
		Pluck("spotify_track_id", &ids)
	return ids, result.Error
}

// DeleteByUserAndTrack, belirli bir kullanıcı-parça çifti için etkileşimi siler.
func (r *repository) DeleteByUserAndTrack(userID uint, spotifyTrackID string) error {
	result := r.db.
		Where("user_id = ? AND spotify_track_id = ?", userID, spotifyTrackID).
		Delete(&TrackInteraction{})
	return result.Error
}

// GetCoLikedTracks, collaborative filtering'in temel sorgusudur.
//
// Mantık: Kullanıcının beğendiği parçaları beğenen diğer kullanıcıların
// ayrıca beğendiği parçaları bulur ve co-occurrence sayısına göre sıralar.
//
// SQL:
//
//	SELECT ti2.spotify_track_id, COUNT(DISTINCT ti2.user_id) as co_like_count
//	FROM track_interactions ti1
//	JOIN track_interactions ti2
//	  ON ti1.user_id = ti2.user_id
//	  AND ti1.spotify_track_id != ti2.spotify_track_id
//	WHERE ti1.spotify_track_id IN (?)
//	  AND ti1.interaction_type = 'like'
//	  AND ti2.interaction_type = 'like'
//	  AND ti2.user_id != ?
//	GROUP BY ti2.spotify_track_id
//	ORDER BY co_like_count DESC
//	LIMIT ?
func (r *repository) GetCoLikedTracks(likedIDs []string, excludeUserID uint, limit int) ([]CoLikedResult, error) {
	if len(likedIDs) == 0 {
		return nil, nil
	}
	if limit <= 0 {
		limit = 20
	}

	var results []CoLikedResult
	err := r.db.Raw(`
		SELECT ti2.spotify_track_id, COUNT(DISTINCT ti2.user_id) as co_like_count
		FROM track_interactions ti1
		JOIN track_interactions ti2
		  ON ti1.user_id = ti2.user_id
		  AND ti1.spotify_track_id != ti2.spotify_track_id
		WHERE ti1.spotify_track_id IN (?)
		  AND ti1.interaction_type = 'like'
		  AND ti2.interaction_type = 'like'
		  AND ti2.user_id != ?
		GROUP BY ti2.spotify_track_id
		ORDER BY co_like_count DESC
		LIMIT ?
	`, likedIDs, excludeUserID, limit).Scan(&results).Error

	return results, err
}
