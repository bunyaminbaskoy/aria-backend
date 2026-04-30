package recommendation

// CreateRecommendationInput, orchestrator pipeline'ının yeni bir öneri
// kümesini veritabanına yazmak için kullandığı iç (internal) DTO'dur.
//
// Bu yapı HTTP'den doğrudan alınmaz; Service.CreateFromAI metoduna
// orchestrator tarafından geçilir. Python AI contract'ı ile birebir
// uyumludur (bkz. pkg/aiclient/contract.go RecommendResponse).
type CreateRecommendationInput struct {
	UserID         uint
	MoodID         uint
	AIModelVersion string
	RAGContext     string
	ProcessingMs   int
	Tracks         []TrackInput
}

// TrackInput, AI servisinden dönen tek bir parça önerisini temsil eder.
// Position alanı orchestrator tarafından sırayla atanır (0, 1, 2, ...).
type TrackInput struct {
	SpotifyTrackID string
	Title          string
	Artist         string
	Album          string
	PreviewURL     string
	ExternalURL    string
	DurationMs     int
	RelevanceScore float64
	Reason         string
}
