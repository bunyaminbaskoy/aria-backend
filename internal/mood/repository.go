package mood

import (
	"errors"

	"gorm.io/gorm"
)

// Repository, Mood verilerine erişim arayüzüdür.
// Service katmanı somut GORM implementasyonuna değil bu arayüze bağımlıdır;
// böylece test sırasında mock'lanabilir ve veri kaynağı ileride değişse bile
// üst katmanlar etkilenmez.
type Repository interface {
	Create(mood *Mood) error
	FindByID(id uint) (*Mood, error)
	FindByUserID(userID uint, limit, offset int) ([]Mood, error)
	Update(mood *Mood) error
	UpdateAnalysis(id uint, update AnalysisUpdate) error
}

// repository, Repository arayüzünün GORM tabanlı implementasyonudur.
// Dışa açık değildir; sadece NewRepository üzerinden erişilir.
type repository struct {
	db *gorm.DB
}

// NewRepository, GORM bağlantısını alarak yeni bir Mood repository üretir.
func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

// Create, veritabanına yeni bir Mood kaydı ekler.
// Çağıran taraf genellikle yalnızca UserID ve RawText'i doldurur;
// AI alanları sonradan UpdateAnalysis ile yazılır.
func (r *repository) Create(mood *Mood) error {
	return r.db.Create(mood).Error
}

// FindByID, birincil anahtara göre tek bir Mood kaydı getirir.
// Kayıt bulunamazsa (nil, nil) döner — hata değil.
func (r *repository) FindByID(id uint) (*Mood, error) {
	var m Mood
	result := r.db.First(&m, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &m, nil
}

// FindByUserID, belirli bir kullanıcıya ait Mood kayıtlarını en yeniden
// eskiye sıralı şekilde döner. Sayfalama için limit/offset kullanılır.
// limit <= 0 ise varsayılan olarak 20 kullanılır.
func (r *repository) FindByUserID(userID uint, limit, offset int) ([]Mood, error) {
	if limit <= 0 {
		limit = 20
	}
	var moods []Mood
	result := r.db.
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&moods)
	return moods, result.Error
}

// Update, mevcut bir Mood kaydının tüm alanlarını günceller.
// Kısmi güncelleme için UpdateAnalysis tercih edilmelidir.
func (r *repository) Update(mood *Mood) error {
	return r.db.Save(mood).Error
}

// UpdateAnalysis, orchestrator pipeline'ı tarafından AI sonuçlarını
// var olan bir Mood kaydına yazmak için kullanılır.
//
// Yalnızca AI'nın doldurduğu alanlar ve Status güncellenir; RawText,
// UserID gibi orijinal alanlar korunur. Tek bir UPDATE sorgusuyla
// atomik olarak çalışır.
func (r *repository) UpdateAnalysis(id uint, update AnalysisUpdate) error {
	updates := map[string]interface{}{
		"sentiment_label":  update.SentimentLabel,
		"dominant_emotion": update.DominantEmotion,
		"valence":          update.Valence,
		"arousal":          update.Arousal,
		"energy":           update.Energy,
		"emotion_scores":   update.EmotionScores,
		"language":         update.Language,
		"ai_model_version": update.AIModelVersion,
		"processing_ms":    update.ProcessingMs,
		"status":           StatusAnalyzed,
	}
	return r.db.Model(&Mood{}).Where("id = ?", id).Updates(updates).Error
}
