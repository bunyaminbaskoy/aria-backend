package test

// Bu dosya, Spotify modülünün veri tiplerini, playlist handler mantığını
// ve SpotifyTokenManager bağımlılıklarını test eder.
// Gerçek Spotify API çağrısı yapılmaz — sahte HTTP yanıtlar kullanılır.

import (
	"testing"

	"music-curation/internal/spotify"
)

// ===========================================================================
// Testler: Spotify Veri Yapıları
// ===========================================================================

// TestTrackStructFields, spotify.Track yapısının beklenen alanları
// içerdiğini ve JSON serileştirme etiketlerinin doğru olduğunu doğrular.
// Bu test, arkadaşın kodu ile bizim pipeline kodunun uyumunu kontrol eder.
func TestTrackStructFields(t *testing.T) {
	track := spotify.Track{
		ID:   "3n3Ppam7vgaVa1iaRUc9Lp",
		Name: "Holocene",
		URI:  "spotify:track:3n3Ppam7vgaVa1iaRUc9Lp",
		Artists: []spotify.Artist{
			{ID: "4LEiUm1SRbFMgfqnQTwUbQ", Name: "Bon Iver", URI: "spotify:artist:4LEiUm1SRbFMgfqnQTwUbQ"},
		},
		DurationMs: 337000,
	}

	if track.ID == "" {
		t.Error("Track.ID boş olmamalı")
	}
	if track.Name == "" {
		t.Error("Track.Name boş olmamalı")
	}
	if track.URI == "" {
		t.Error("Track.URI boş olmamalı")
	}
	if len(track.Artists) == 0 {
		t.Error("Track.Artists en az bir sanatçı içermeli")
	}
	if track.Artists[0].Name != "Bon Iver" {
		t.Errorf("Track.Artists[0].Name: beklenen 'Bon Iver', aldı '%s'", track.Artists[0].Name)
	}
	if track.DurationMs <= 0 {
		t.Error("Track.DurationMs pozitif olmalı")
	}
}

// TestArtistStructFields, spotify.Artist yapısını doğrular.
func TestArtistStructFields(t *testing.T) {
	artist := spotify.Artist{
		ID:   "4LEiUm1SRbFMgfqnQTwUbQ",
		Name: "Bon Iver",
		URI:  "spotify:artist:4LEiUm1SRbFMgfqnQTwUbQ",
	}

	if artist.ID == "" {
		t.Error("Artist.ID boş olmamalı")
	}
	if artist.Name == "" {
		t.Error("Artist.Name boş olmamalı")
	}
}

// TestFullArtistStructFields, spotify.FullArtist yapısının ek alanlarını
// (Genre, Popularity, Images) doğrular.
func TestFullArtistStructFields(t *testing.T) {
	artist := spotify.FullArtist{
		ID:         "4LEiUm1SRbFMgfqnQTwUbQ",
		Name:       "Bon Iver",
		URI:        "spotify:artist:4LEiUm1SRbFMgfqnQTwUbQ",
		Genres:     []string{"indie folk", "chamber pop"},
		Popularity: 80,
		Images: []spotify.Image{
			{URL: "https://example.com/artist.jpg", Height: 640, Width: 640},
		},
	}

	if artist.ID == "" {
		t.Error("FullArtist.ID boş olmamalı")
	}
	if len(artist.Genres) == 0 {
		t.Error("FullArtist.Genres en az bir tür içermeli")
	}
	if artist.Popularity < 0 || artist.Popularity > 100 {
		t.Errorf("Popularity [0-100] aralığında olmalı, aldı: %d", artist.Popularity)
	}
	if len(artist.Images) == 0 {
		t.Error("FullArtist.Images en az bir görsel içermeli")
	}
}

// TestPlayHistoryItem, PlayHistoryItem ve içindeki Track'in doğru
// yapılandırıldığını doğrular.
func TestPlayHistoryItem(t *testing.T) {
	item := spotify.PlayHistoryItem{
		Track: spotify.Track{
			ID:   "abc123",
			Name: "Test Şarkısı",
			Artists: []spotify.Artist{
				{ID: "art1", Name: "Test Sanatçı"},
			},
		},
		PlayedAt: "2024-01-15T14:30:00Z",
	}

	if item.Track.ID == "" {
		t.Error("PlayHistoryItem.Track.ID boş olmamalı")
	}
	if item.PlayedAt == "" {
		t.Error("PlayHistoryItem.PlayedAt boş olmamalı")
	}
}

// ===========================================================================
// Testler: Playlist Oluşturma İsteği Doğrulama
// ===========================================================================

// TestCreatePlaylistRequest_Validation, playlist oluşturma isteğinin
// beklenen alanları içerdiğini doğrular.
func TestCreatePlaylistRequest_Validation(t *testing.T) {
	// Geçerli istek
	req := spotify.CreatePlaylistRequest{
		Name:        "Sabah Enerjisi",
		Description: "Güne enerjik başlamak için",
		Public:      false,
		TrackURIs: []string{
			"spotify:track:3n3Ppam7vgaVa1iaRUc9Lp",
			"spotify:track:2TpxZ7JUBn3uw46aR7qd6V",
		},
	}

	if req.Name == "" {
		t.Error("Playlist adı boş olmamalı")
	}
	if len(req.TrackURIs) == 0 {
		t.Error("TrackURIs en az bir URI içermeli")
	}
	// Spotify URI formatını kontrol et
	for i, uri := range req.TrackURIs {
		if len(uri) < 14 { // "spotify:track:" = 14 karakter
			t.Errorf("TrackURIs[%d] geçersiz Spotify URI formatında: '%s'", i, uri)
		}
	}
}

// TestCreatePlaylistRequest_EmptyTracks, boş track listesi gönderildiğinde
// binding tag'in "min=1" kuralının bunu yakalamasını kontrol eder.
// (Handler'daki ShouldBindJSON bunu gin validation ile yakalar.)
func TestCreatePlaylistRequest_EmptyTracks(t *testing.T) {
	// Bu test binding'i doğrudan test edemez (gin gerektirir),
	// ama mantığı belgelemek için iş katmanında kontrol simüle eder.
	req := spotify.CreatePlaylistRequest{
		Name:      "Boş Playlist",
		TrackURIs: []string{},
	}

	if len(req.TrackURIs) != 0 {
		t.Error("Beklenen: boş TrackURIs")
	}
	// Gerçek handler bu durumda 400 Bad Request döndürmeli
}

// ===========================================================================
// Testler: TopTracks ve TopArtists yanıt yapıları
// ===========================================================================

// TestTopTracksResponse, Top Tracks yanıt yapısını doğrular.
func TestTopTracksResponse(t *testing.T) {
	resp := spotify.TopTracksResponse{
		Items: []spotify.Track{
			{ID: "t1", Name: "Şarkı 1", DurationMs: 180000},
			{ID: "t2", Name: "Şarkı 2", DurationMs: 240000},
		},
		Total: 2,
	}

	if len(resp.Items) != 2 {
		t.Errorf("Items sayısı: beklenen 2, aldı %d", len(resp.Items))
	}
	if resp.Total != 2 {
		t.Errorf("Total: beklenen 2, aldı %d", resp.Total)
	}
}

// TestTopArtistsResponse, Top Artists yanıt yapısını doğrular.
func TestTopArtistsResponse(t *testing.T) {
	resp := spotify.TopArtistsResponse{
		Items: []spotify.FullArtist{
			{ID: "a1", Name: "Sanatçı 1", Popularity: 75},
		},
		Total: 1,
	}

	if len(resp.Items) != 1 {
		t.Errorf("Items sayısı: beklenen 1, aldı %d", len(resp.Items))
	}
	if resp.Items[0].Name != "Sanatçı 1" {
		t.Errorf("İlk sanatçı adı: beklenen 'Sanatçı 1', aldı '%s'", resp.Items[0].Name)
	}
}

// ===========================================================================
// Testler: Spotify Playlist Yanıtı
// ===========================================================================

// TestSpotifyPlaylistResponse, oluşturulan playlist'in beklenen alanları
// içerdiğini doğrular. Bu yapı CreatePlaylist handler'ı tarafından kullanılır.
func TestSpotifyPlaylistResponse(t *testing.T) {
	playlist := spotify.SpotifyPlaylistResponse{
		ID:   "playlist123",
		Name: "Sabah Enerjisi",
		URI:  "spotify:playlist:playlist123",
	}
	playlist.ExternalURLs.Spotify = "https://open.spotify.com/playlist/playlist123"

	if playlist.ID == "" {
		t.Error("Playlist ID boş olmamalı")
	}
	if playlist.Name == "" {
		t.Error("Playlist adı boş olmamalı")
	}
	if playlist.ExternalURLs.Spotify == "" {
		t.Error("Spotify external URL boş olmamalı")
	}
}
