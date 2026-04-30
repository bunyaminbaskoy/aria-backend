package recommendation

import (
	"errors"
)

// Service, Recommendation modülünün iş kuralı katmanıdır.
// Orchestrator pipeline'ı ve HTTP handler'ı bu yapıya bağımlıdır.
type Service struct {
	repo Repository
}

// NewService, Repository'i alarak yeni bir Recommendation service üretir.
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// ErrNoTracks, AI servisi boş bir parça listesi döndürdüğünde fırlatılır.
// Boş öneri kaydı yaratmak veritabanını kirletir, bu yüzden service
// katmanında erkenden engellenir.
var ErrNoTracks = errors.New("öneri için en az bir parça gereklidir")

// CreateFromAI, orchestrator pipeline'ı tarafından AI servisinden dönen
// sonuçlarla yeni bir Recommendation + Track kümesi yazmak için
// çağrılır. Bu, modüller arası iletişimin direkt Go fonksiyon çağrısı
// olarak yapıldığı bir noktadır — ağ üzerinden değil aynı süreç içinde.
//
// Position alanı, gelen Tracks slice'ının sırasına göre 0'dan
// itibaren otomatik olarak atanır.
func (s *Service) CreateFromAI(input CreateRecommendationInput) (*Recommendation, error) {
	if len(input.Tracks) == 0 {
		return nil, ErrNoTracks
	}

	tracks := make([]RecommendedTrack, 0, len(input.Tracks))
	for i, t := range input.Tracks {
		tracks = append(tracks, RecommendedTrack{
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

	rec := &Recommendation{
		UserID:         input.UserID,
		MoodID:         input.MoodID,
		AIModelVersion: input.AIModelVersion,
		RAGContext:     input.RAGContext,
		ProcessingMs:   input.ProcessingMs,
		Status:         StatusReady,
		Tracks:         tracks,
	}

	if err := s.repo.Create(rec); err != nil {
		return nil, err
	}
	return rec, nil
}

// CreatePending, orchestrator henüz AI cevabını almadan önce "yer
// tutucu" bir Recommendation kaydı oluşturmak isterse kullanılır.
// Asenkron pipeline akışını desteklemek için tasarlandı; senkron
// akışta CreateFromAI doğrudan çağrılır.
func (s *Service) CreatePending(userID, moodID uint) (*Recommendation, error) {
	rec := &Recommendation{
		UserID: userID,
		MoodID: moodID,
		Status: StatusPending,
	}
	if err := s.repo.Create(rec); err != nil {
		return nil, err
	}
	return rec, nil
}

// MarkFailed, AI çağrısı başarısız olduğunda kaydı "failed" olarak
// işaretler. Kullanıcıya frontend'de "tekrar dene" butonu göstermek
// için bu durum okunur.
func (s *Service) MarkFailed(id uint) error {
	return s.repo.UpdateStatus(id, StatusFailed)
}

// GetByID, tek bir öneri kümesini Tracks ile birlikte döner.
// Kayıt yoksa (nil, nil) döner.
func (s *Service) GetByID(id uint) (*Recommendation, error) {
	return s.repo.FindByID(id)
}

// GetUserRecommendations, kullanıcının kendi öneri geçmişini sayfalı
// olarak döner. Liste görünümü olduğu için tracks dahil edilmez.
func (s *Service) GetUserRecommendations(userID uint, limit, offset int) ([]Recommendation, error) {
	return s.repo.FindByUserID(userID, limit, offset)
}

// GetByMoodID, belirli bir Mood için üretilmiş tüm önerileri döner.
// Tracks preload'lı gelir.
func (s *Service) GetByMoodID(moodID uint) ([]Recommendation, error) {
	return s.repo.FindByMoodID(moodID)
}
