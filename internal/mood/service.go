package mood

import (
	"errors"
	"strings"
)

// Service, Mood modülünün iş kuralı (business logic) katmanıdır.
// HTTP/orchestrator gibi üst katmanlar doğrudan repository'e değil
// bu yapıya bağımlıdır.
type Service struct {
	repo Repository
}

// NewService, Repository'i alarak yeni bir Mood service üretir.
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// ErrEmptyText, kullanıcı boş veya yalnızca boşluk içeren bir metin
// gönderdiğinde döner. Handler bu hatayı 400 Bad Request'e çevirir.
var ErrEmptyText = errors.New("mood metni boş olamaz")

// CreateRawMood, kullanıcının girdiği ham metin için yeni bir Mood
// kaydı yaratır. Kayıt başlangıçta "pending" durumundadır; AI analizi
// orchestrator tarafından sonradan UpdateAnalysis ile uygulanır.
//
// Bu metot bilinçli olarak AI servisini çağırmaz — modüller arası
// bağımlılığı tek yöne (orchestrator → mood) tutmak için. Modular
// monolith içinde mood modülü AI'dan habersizdir.
func (s *Service) CreateRawMood(userID uint, text string) (*Mood, error) {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return nil, ErrEmptyText
	}

	m := &Mood{
		UserID:  userID,
		RawText: trimmed,
		Status:  StatusPending,
	}
	if err := s.repo.Create(m); err != nil {
		return nil, err
	}
	return m, nil
}

// GetMoodByID, ID ile tek bir Mood kaydı döner. Kayıt yoksa (nil, nil)
// döner; handler bunu 404'e çevirir.
func (s *Service) GetMoodByID(id uint) (*Mood, error) {
	return s.repo.FindByID(id)
}

// GetUserMoods, bir kullanıcıya ait son ruh hali kayıtlarını sayfalı
// olarak döner. limit <= 0 verilirse repository varsayılanını kullanır.
func (s *Service) GetUserMoods(userID uint, limit, offset int) ([]Mood, error) {
	return s.repo.FindByUserID(userID, limit, offset)
}

// UpdateAnalysis, orchestrator pipeline'ı tarafından AI sonuçlarını
// var olan bir Mood kaydına işlemek için çağrılır. Bu metot, modüller
// arası iletişimin direkt Go fonksiyon çağrısı olarak yapıldığı tek
// noktadır; ağ üzerinden değil aynı süreç içinde çalışır.
func (s *Service) UpdateAnalysis(moodID uint, update AnalysisUpdate) error {
	return s.repo.UpdateAnalysis(moodID, update)
}

// MarkFailed, AI çağrısı başarısız olduğunda Mood kaydının Status alanını
// "failed" olarak işaretler. RawText korunduğu için kullanıcı tekrar
// deneyebilir.
func (s *Service) MarkFailed(moodID uint) error {
	m, err := s.repo.FindByID(moodID)
	if err != nil {
		return err
	}
	if m == nil {
		return nil
	}
	m.Status = StatusFailed
	return s.repo.Update(m)
}
