package mood

// CreateMoodRequest, kullanıcının ruh hali metnini göndermek için
// kullandığı request gövdesini temsil eder.
//
// Örnek:
//
//	{
//	  "text": "Bugün çok yorgunum ama mutluyum"
//	}
type CreateMoodRequest struct {
	Text string `json:"text" binding:"required,min=1,max=2000"`
}

// AnalysisUpdate, orchestrator pipeline'ı tarafından AI servisinden
// dönen sentiment sonuçlarını mevcut bir Mood kaydına uygulamak için
// kullanılan iç (internal) DTO'dur.
//
// Bu struct dışarıya (HTTP) açılmaz; yalnızca Service.UpdateAnalysis
// metoduna geçilir. Python AI contract'ı ile birebir uyumlu olacak
// şekilde tasarlanmıştır (bkz. pkg/aiclient/contract.go).
type AnalysisUpdate struct {
	SentimentLabel  string
	DominantEmotion string
	Valence         float64
	Arousal         float64
	Energy          float64
	EmotionScores   []byte // ham JSON; doğrudan datatypes.JSON kolonuna yazılır
	Language        string
	AIModelVersion  string
	ProcessingMs    int
}
