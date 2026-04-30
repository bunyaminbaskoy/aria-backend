package recommendation

import (
	"errors"

	"gorm.io/gorm"
)

// Repository, Recommendation ve RecommendedTrack verilerine erişim
// arayüzüdür. Service katmanı bu arayüze bağımlıdır; somut GORM
// implementasyonu test sırasında mock'lanabilir.
type Repository interface {
	Create(rec *Recommendation) error
	FindByID(id uint) (*Recommendation, error)
	FindByUserID(userID uint, limit, offset int) ([]Recommendation, error)
	FindByMoodID(moodID uint) ([]Recommendation, error)
	UpdateStatus(id uint, status string) error
}

// repository, Repository arayüzünün GORM tabanlı implementasyonudur.
type repository struct {
	db *gorm.DB
}

// NewRepository, GORM bağlantısını alarak yeni bir Recommendation
// repository üretir.
func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

// Create, parent Recommendation ve tüm child Track kayıtlarını tek bir
// transaction içinde yazar. GORM'un has-many ilişkisi sayesinde
// rec.Tracks alanı dolu gelirse otomatik olarak birlikte INSERT edilir;
// transaction sırasında bir parça yazılamazsa parent da rollback edilir.
//
// Bu atomiklik, "yarım yamalak öneri kümesi" gibi tutarsız durumları engeller.
func (r *repository) Create(rec *Recommendation) error {
	return r.db.Create(rec).Error
}

// FindByID, ID ile tek bir Recommendation kaydını ve ona bağlı tüm
// RecommendedTrack'leri Position'a göre sıralı şekilde getirir.
// Kayıt bulunamazsa (nil, nil) döner.
func (r *repository) FindByID(id uint) (*Recommendation, error) {
	var rec Recommendation
	result := r.db.
		Preload("Tracks", func(db *gorm.DB) *gorm.DB {
			return db.Order("position ASC")
		}).
		First(&rec, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &rec, nil
}

// FindByUserID, bir kullanıcının önerilerini en yeniden eskiye sıralı
// olarak döner. Liste görünümü için tracks Preload edilmez (performans);
// detay görünümünde FindByID ile preload'lı versiyon kullanılır.
func (r *repository) FindByUserID(userID uint, limit, offset int) ([]Recommendation, error) {
	if limit <= 0 {
		limit = 20
	}
	var recs []Recommendation
	result := r.db.
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&recs)
	return recs, result.Error
}

// FindByMoodID, belirli bir Mood için üretilmiş tüm önerileri döner.
// Tracks preload edilir; tipik olarak 1-2 kayıt döner.
func (r *repository) FindByMoodID(moodID uint) ([]Recommendation, error) {
	var recs []Recommendation
	result := r.db.
		Where("mood_id = ?", moodID).
		Preload("Tracks", func(db *gorm.DB) *gorm.DB {
			return db.Order("position ASC")
		}).
		Order("created_at DESC").
		Find(&recs)
	return recs, result.Error
}

// UpdateStatus, yalnızca Status alanını günceller. Asenkron pipeline
// "pending" -> "ready" / "failed" geçişleri için kullanılır.
func (r *repository) UpdateStatus(id uint, status string) error {
	return r.db.Model(&Recommendation{}).
		Where("id = ?", id).
		Update("status", status).Error
}
