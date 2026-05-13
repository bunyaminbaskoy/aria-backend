// Package aiclient, Go backend ile Python (FastAPI) tabanlı AI/RAG
// servisi arasındaki HTTP sözleşmesini (contract) tanımlar.
//
// =============================================================================
//
//	🚨 BU DOSYA, PYTHON EKİBİYLE PAYLAŞILACAK RESMİ SPESİFİKASYONDUR.
//
// =============================================================================
//
// Bu dosyada tanımlanan struct'lar, JSON etiketleriyle (json:"...") birlikte
// Python FastAPI servisinin uygulaması GEREKEN istek/yanıt şemalarını
// belirler. Python ekibi Pydantic modellerini bu yapılarla birebir uyumlu
// olacak şekilde yazmalıdır.
//
// Alan adları snake_case'tir (Python tarafında doğal okunsun diye).
// Tüm sayısal alanlar JSON number olarak gönderilir (string DEĞİL).
// Nullable/opsiyonel alanlar `omitempty` ile işaretlenmiştir; Python
// tarafında Optional[...] kullanılmalıdır.
//
// =============================================================================
//
// Endpoint'ler:
//
//	POST {AI_SERVICE_URL}/analyze     — sentiment + duygu analizi
//	POST {AI_SERVICE_URL}/recommend   — RAG tabanlı parça önerisi
//
// Her iki endpoint de:
//   - Content-Type: application/json
//   - Hata durumunda HTTP 4xx/5xx + ErrorResponse JSON gövdesi döner.
//   - 200 OK üzerinde aşağıdaki ilgili response struct'ı döner.
//
// =============================================================================
package aiclient

// -----------------------------------------------------------------------------
// /analyze — Sentiment ve Duygu Analizi
// -----------------------------------------------------------------------------

// AnalyzeRequest, /analyze endpoint'ine gönderilen istek gövdesidir.
//
// Örnek istek:
//
//	POST /analyze
//	{
//	  "text": "Bugün çok yorgunum ama mutluyum",
//	  "user_id": 42,
//	  "language_hint": "tr"
//	}
type AnalyzeRequest struct {
	// Text, kullanıcının analiz edilecek ham metnidir. Backend tarafında
	// trim edilmiş ve boş olmadığı doğrulanmış halde gönderilir.
	// Maksimum uzunluk: 2000 karakter.
	Text string `json:"text"`

	// UserID, isteği yapan kullanıcının iç (internal) ID'sidir.
	// Python tarafında kişiselleştirme veya log için kullanılabilir.
	UserID uint `json:"user_id"`

	// LanguageHint, opsiyoneldir. Backend dili biliyorsa (örn. kullanıcı
	// profilinden) gönderir; bilmiyorsa boş bırakır ve AI servisi otomatik
	// olarak tespit eder.
	LanguageHint string `json:"language_hint,omitempty"`
}

// AnalyzeResponse, /analyze endpoint'inden dönen başarılı yanıtın
// gövdesidir. Backend bu alanları doğrudan Mood tablosundaki
// karşılık gelen kolonlara yazar.
//
// Örnek yanıt:
//
//	{
//	  "sentiment_label": "bittersweet",
//	  "dominant_emotion": "contentment",
//	  "valence": 0.45,
//	  "arousal": -0.20,
//	  "energy": 0.30,
//	  "emotion_scores": {
//	    "joy": 0.6,
//	    "sadness": 0.2,
//	    "fatigue": 0.7
//	  },
//	  "language": "tr",
//	  "model_version": "aria-sentiment-v1.0.0",
//	  "processing_ms": 145
//	}
type AnalyzeResponse struct {
	// SentimentLabel, üst seviye etiket. Önerilen değerler:
	//   "positive" | "negative" | "neutral" | "mixed" | "bittersweet"
	// Liste kapalı değil — Python ekibi gerekirse genişletebilir.
	SentimentLabel string `json:"sentiment_label"`

	// DominantEmotion, EmotionScores içindeki en yüksek skora sahip
	// duygunun adıdır. Frontend'de tek kelimelik özet için kullanılır.
	DominantEmotion string `json:"dominant_emotion"`

	// Valence, [-1.0, +1.0] aralığında. -1 negatif, +1 pozitif duygulanım.
	// Russell's circumplex model'inden esinlenilmiştir.
	Valence float64 `json:"valence"`

	// Arousal, [-1.0, +1.0] aralığında. -1 sakin, +1 heyecanlı/uyarılmış.
	Arousal float64 `json:"arousal"`

	// Energy, [0.0, 1.0] aralığında. Müzik öneri tarafında kullanılır;
	// Spotify'ın "energy" özelliğine yakın bir kavram.
	Energy float64 `json:"energy"`

	// EmotionScores, çok-etiketli (multi-label) duygu skorları.
	// Anahtarlar serbest stringtir (joy, sadness, anger, fear, surprise,
	// disgust, fatigue, calm, ...). Skorlar [0.0, 1.0] aralığında olmalıdır.
	// Boş dönmesi de geçerlidir.
	EmotionScores map[string]float64 `json:"emotion_scores,omitempty"`

	// Language, ISO 639-1 dil kodu (örn. "tr", "en", "es").
	Language string `json:"language,omitempty"`

	// ModelVersion, bu cevabı üreten model sürümü. Reprocessing ve A/B
	// karşılaştırma için backend'de saklanır.
	ModelVersion string `json:"model_version"`

	// ProcessingMs, AI tarafının harcadığı süre (milisaniye). Network
	// latency hariçtir.
	ProcessingMs int `json:"processing_ms"`
}

// -----------------------------------------------------------------------------
// /recommend — RAG Tabanlı Parça Önerisi
// -----------------------------------------------------------------------------

// RecommendRequest, /recommend endpoint'ine gönderilen istek gövdesidir.
//
// AI servisi, MoodSnapshot içindeki sentiment vektörünü ve opsiyonel
// Context bilgilerini kullanarak RAG (Retrieval-Augmented Generation)
// üzerinden ilgili parçaları döner.
//
// Örnek istek:
//
//	POST /recommend
//	{
//	  "user_id": 42,
//	  "mood_id": 17,
//	  "mood": {
//	    "sentiment_label": "bittersweet",
//	    "dominant_emotion": "contentment",
//	    "valence": 0.45,
//	    "arousal": -0.20,
//	    "energy": 0.30
//	  },
//	  "limit": 20,
//	  "context": {
//	    "preferred_genres": ["indie", "lo-fi"],
//	    "exclude_track_ids": ["spotify:track:abc123"],
//	    "language": "tr"
//	  }
//	}
type RecommendRequest struct {
	// UserID, isteği yapan kullanıcının iç ID'si.
	UserID uint `json:"user_id"`

	// MoodID, ilişkili Mood kaydının ID'si. Python tarafında öneriyi
	// loglarken referans olarak kullanılabilir.
	MoodID uint `json:"mood_id"`

	// Mood, sentiment analizi sonucunun snapshot'ıdır. Bu sayede AI
	// servisi tekrar analiz yapmak zorunda kalmaz.
	Mood MoodSnapshot `json:"mood"`

	// Limit, döndürülecek maksimum parça sayısı. Backend varsayılan
	// olarak 20 gönderir; Python en fazla 50'yi desteklemelidir.
	Limit int `json:"limit"`

	// Context, opsiyonel kişiselleştirme bilgileridir.
	Context *RecommendContext `json:"context,omitempty"`
}

// MoodSnapshot, Recommendation isteğine gömülen sentiment özetidir.
// Alanlar AnalyzeResponse'tan birebir kopyalanır.
type MoodSnapshot struct {
	SentimentLabel  string  `json:"sentiment_label"`
	DominantEmotion string  `json:"dominant_emotion"`
	Valence         float64 `json:"valence"`
	Arousal         float64 `json:"arousal"`
	Energy          float64 `json:"energy"`
}

// RecommendContext, kullanıcının geçmişine ve tercihlerine dair
// opsiyonel ipuçlarıdır. Tüm alanlar nullable'dır.
type RecommendContext struct {
	// PreferredGenres, kullanıcının tercih ettiği müzik türleri.
	// Örn: ["indie", "jazz", "ambient"].
	PreferredGenres []string `json:"preferred_genres,omitempty"`

	// ExcludeTrackIDs, kullanıcının daha önce dinlediği veya
	// dışlamak istediği parçaların Spotify ID'leri.
	ExcludeTrackIDs []string `json:"exclude_track_ids,omitempty"`

	// Language, kullanıcı arayüzünün dili (öneri açıklamaları
	// "reason" alanında bu dilde dönsün diye).
	Language string `json:"language,omitempty"`

	// LikedTrackIDs, kullanıcının daha önce beğendiği parçaların
	// Spotify ID'leri. Collaborative filtering sinyali olarak
	// Python ML servisine iletilir.
	LikedTrackIDs []string `json:"liked_track_ids,omitempty"`

	// CollabTrackIDs, collaborative filtering (co-occurrence) sorgusu
	// sonucunda bulunan parça ID'leridir. Benzer zevklere sahip diğer
	// kullanıcıların beğendiği parçaları temsil eder.
	CollabTrackIDs []string `json:"collab_track_ids,omitempty"`
}

// RecommendResponse, /recommend endpoint'inden dönen başarılı yanıtın
// gövdesidir. Backend bu yapıyı Recommendation + RecommendedTrack
// kayıtlarına dönüştürerek persist eder.
//
// Örnek yanıt:
//
//	{
//	  "model_version": "aria-rag-v1.0.0",
//	  "rag_context": "User mood matches calm contentment dimension; retrieved 24 tracks from corpus.",
//	  "processing_ms": 320,
//	  "tracks": [
//	    {
//	      "spotify_track_id": "3n3Ppam7vgaVa1iaRUc9Lp",
//	      "title": "Holocene",
//	      "artist": "Bon Iver",
//	      "album": "Bon Iver, Bon Iver",
//	      "preview_url": "https://p.scdn.co/mp3-preview/...",
//	      "external_url": "https://open.spotify.com/track/3n3Ppam7vgaVa1iaRUc9Lp",
//	      "duration_ms": 337000,
//	      "relevance_score": 0.92,
//	      "reason": "Sakin vokaller ve dingin atmosfer, içe dönük mutluluk hissini güçlendiriyor."
//	    }
//	  ]
//	}
type RecommendResponse struct {
	// ModelVersion, RAG modelinin sürümü.
	ModelVersion string `json:"model_version"`

	// RAGContext, AI servisinin öneri sürecinde kullandığı bağlam metni.
	// Açıklanabilirlik (explainability) ve debug için backend'de saklanır.
	// Boş gelmesi geçerlidir.
	RAGContext string `json:"rag_context,omitempty"`

	// ProcessingMs, AI tarafının harcadığı süre (milisaniye).
	ProcessingMs int `json:"processing_ms"`

	// Tracks, önerilen parçalar. Sıralama önemlidir — backend bu sırayı
	// Position alanına yazar (0'dan başlayarak). En alakalı parça önce
	// gelmelidir. Boş liste dönmesi yerine HTTP 422 + ErrorResponse
	// dönülmesi tercih edilir.
	Tracks []TrackSuggestion `json:"tracks"`
}

// TrackSuggestion, AI'nın önerdiği tekil bir parçayı temsil eder.
type TrackSuggestion struct {
	// SpotifyTrackID, Spotify katalog ID'si. AI servisi bu ID'yi
	// güvenilir biçimde üretemiyorsa boş bırakabilir; backend daha
	// sonra Spotify search ile çözümler.
	SpotifyTrackID string `json:"spotify_track_id,omitempty"`

	// Title, parça adı. Zorunludur.
	Title string `json:"title"`

	// Artist, ana sanatçı adı (virgülle birden fazla yazılabilir
	// ama tek string olarak gelmelidir). Zorunludur.
	Artist string `json:"artist"`

	// Album, albüm adı. Opsiyoneldir.
	Album string `json:"album,omitempty"`

	// PreviewURL, Spotify'ın 30 saniyelik mp3 preview linki. Opsiyonel.
	PreviewURL string `json:"preview_url,omitempty"`

	// ExternalURL, parçanın Spotify uygulamasında açıldığı URL.
	// Genellikle "https://open.spotify.com/track/{id}" formatındadır.
	ExternalURL string `json:"external_url,omitempty"`

	// DurationMs, parça süresi (milisaniye). Bilinmiyorsa 0 gönderilir.
	DurationMs int `json:"duration_ms,omitempty"`

	// RelevanceScore, [0.0, 1.0] aralığında. AI'nın bu parçanın ruh
	// hali ile ne kadar örtüştüğüne dair güven skoru.
	RelevanceScore float64 `json:"relevance_score"`

	// Reason, frontend'de "neden bu şarkı?" ipucu için kullanılan
	// kısa, kullanıcı dostu açıklama. Dilin RecommendContext.Language
	// ile uyumlu olması beklenir. Opsiyoneldir.
	Reason string `json:"reason,omitempty"`
}

// -----------------------------------------------------------------------------
// Hata Sözleşmesi
// -----------------------------------------------------------------------------

// ErrorResponse, AI servisinin tüm hata durumlarında dönmesi beklenen
// standart yanıt gövdesidir. HTTP status code'u ile birlikte kullanılır.
//
// Önerilen status kodları:
//
//	400 Bad Request           — request gövdesi geçersiz
//	422 Unprocessable Entity  — istek geçerli ama içerik işlenemedi
//	                            (örn. boş metin, desteklenmeyen dil)
//	500 Internal Server Error — model yüklenemedi, beklenmeyen hata
//	503 Service Unavailable   — model şu anda kullanılamıyor / overload
//
// Örnek:
//
//	HTTP/1.1 422 Unprocessable Entity
//	{
//	  "error": "validation_failed",
//	  "message": "text alanı 2000 karakteri aşamaz",
//	  "details": {
//	    "field": "text",
//	    "max_length": 2000
//	  }
//	}
type ErrorResponse struct {
	// Error, makine tarafından okunabilen kısa hata kodu (snake_case).
	// Örn: "validation_failed", "model_unavailable", "rate_limited".
	Error string `json:"error"`

	// Message, insan tarafından okunabilen açıklama.
	Message string `json:"message"`

	// Details, opsiyonel ek bağlam (alan adı, limit, vb.).
	Details map[string]interface{} `json:"details,omitempty"`
}
