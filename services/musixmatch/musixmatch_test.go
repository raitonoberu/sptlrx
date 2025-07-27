package musixmatch

import (
	"testing"

	"github.com/raitonoberu/sptlrx/lyrics"
)

func TestNew(t *testing.T) {
	client := New("test-api-key")

	if client == nil {
		t.Fatal("New should return a non-nil client")
	}

	if client.apiKey != "test-api-key" {
		t.Errorf("Expected API key 'test-api-key', got '%s'", client.apiKey)
	}

	if client.baseURL != baseURL {
		t.Errorf("Expected baseURL '%s', got '%s'", baseURL, client.baseURL)
	}
}

func TestLyricsEmptyAPIKey(t *testing.T) {
	client := New("")

	_, err := client.Lyrics("", "Test Song")
	if err == nil {
		t.Error("Expected error for empty API key")
	}

	if err.Error() != "MusixMatch API key is required" {
		t.Errorf("Expected 'MusixMatch API key is required', got '%s'", err.Error())
	}
}

func TestLyricsEmptyQuery(t *testing.T) {
	client := New("test-key")

	_, err := client.Lyrics("", "")
	if err == nil {
		t.Error("Expected error for empty query")
	}

	if err.Error() != "query cannot be empty" {
		t.Errorf("Expected 'query cannot be empty', got '%s'", err.Error())
	}
}

func TestParseQuery(t *testing.T) {
	testCases := []struct {
		name           string
		query          string
		expectedArtist string
		expectedTrack  string
	}{
		{
			name:           "Artist - Track format",
			query:          "The Beatles - Hey Jude",
			expectedArtist: "The Beatles",
			expectedTrack:  "Hey Jude",
		},
		{
			name:           "Track only",
			query:          "Bohemian Rhapsody",
			expectedArtist: "",
			expectedTrack:  "Bohemian Rhapsody",
		},
		{
			name:           "Multiple dashes",
			query:          "Artist - Track - Remix",
			expectedArtist: "Artist",
			expectedTrack:  "Track - Remix",
		},
		{
			name:           "Empty query",
			query:          "",
			expectedArtist: "",
			expectedTrack:  "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			artist, track := parseQuery(tc.query)
			if artist != tc.expectedArtist {
				t.Errorf("Expected artist '%s', got '%s'", tc.expectedArtist, artist)
			}
			if track != tc.expectedTrack {
				t.Errorf("Expected track '%s', got '%s'", tc.expectedTrack, track)
			}
		})
	}
}

func TestParseTimeString(t *testing.T) {
	testCases := []struct {
		name     string
		timeStr  string
		expected int
	}{
		{
			name:     "Minutes and seconds",
			timeStr:  "03:45",
			expected: 225000, // 3*60*1000 + 45*1000
		},
		{
			name:     "With centiseconds",
			timeStr:  "02:30.50",
			expected: 150500, // 2*60*1000 + 30*1000 + 50*10
		},
		{
			name:     "Zero time",
			timeStr:  "00:00",
			expected: 0,
		},
		{
			name:     "Invalid format",
			timeStr:  "invalid",
			expected: -1,
		},
		{
			name:     "Single number",
			timeStr:  "123",
			expected: -1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := parseTimeString(tc.timeStr)
			if result != tc.expected {
				t.Errorf("Expected %d, got %d", tc.expected, result)
			}
		})
	}
}

func TestParseLRCSubtitles(t *testing.T) {
	client := New("test-key")

	testCases := []struct {
		name        string
		lrcContent  string
		expectedLen int
		firstLine   string
		firstTime   int
	}{
		{
			name: "Basic LRC format",
			lrcContent: `[00:12.50]Line one
[00:17.20]Line two
[00:25.00]Line three`,
			expectedLen: 3,
			firstLine:   "Line one",
			firstTime:   12500,
		},
		{
			name: "With empty lines",
			lrcContent: `[00:12.50]Line one

[00:17.20]Line two`,
			expectedLen: 2,
			firstLine:   "Line one",
			firstTime:   12500,
		},
		{
			name:        "No valid lines",
			lrcContent:  "Invalid content without timestamps",
			expectedLen: 0,
		},
		{
			name:        "Empty content",
			lrcContent:  "",
			expectedLen: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			lines, err := client.parseLRCSubtitles(tc.lrcContent)
			if err != nil {
				t.Errorf("Expected no error, got %v", err)
			}

			if len(lines) != tc.expectedLen {
				t.Errorf("Expected %d lines, got %d", tc.expectedLen, len(lines))
			}

			if tc.expectedLen > 0 {
				if lines[0].Words != tc.firstLine {
					t.Errorf("Expected first line '%s', got '%s'", tc.firstLine, lines[0].Words)
				}
				if lines[0].Time != tc.firstTime {
					t.Errorf("Expected first time %d, got %d", tc.firstTime, lines[0].Time)
				}
			}
		})
	}
}

func TestTrackStruct(t *testing.T) {
	track := Track{
		TrackID:       12345,
		TrackName:     "Test Song",
		ArtistName:    "Test Artist",
		AlbumName:     "Test Album",
		TrackLength:   240,
		HasLyrics:     1,
		HasSubtitles:  1,
		TrackShareURL: "https://musixmatch.com/test",
	}

	if track.TrackID != 12345 {
		t.Errorf("Expected TrackID 12345, got %d", track.TrackID)
	}

	if track.TrackName != "Test Song" {
		t.Errorf("Expected TrackName 'Test Song', got '%s'", track.TrackName)
	}

	if track.HasLyrics != 1 {
		t.Errorf("Expected HasLyrics 1, got %d", track.HasLyrics)
	}

	if track.HasSubtitles != 1 {
		t.Errorf("Expected HasSubtitles 1, got %d", track.HasSubtitles)
	}
}

func TestLyricsDataStruct(t *testing.T) {
	lyricsData := LyricsData{
		LyricsID:        67890,
		LyricsBody:      "Test lyrics content",
		LyricsLanguage:  "en",
		LyricsCopyright: "Test Copyright",
	}

	if lyricsData.LyricsID != 67890 {
		t.Errorf("Expected LyricsID 67890, got %d", lyricsData.LyricsID)
	}

	if lyricsData.LyricsBody != "Test lyrics content" {
		t.Errorf("Expected LyricsBody 'Test lyrics content', got '%s'", lyricsData.LyricsBody)
	}

	if lyricsData.LyricsLanguage != "en" {
		t.Errorf("Expected LyricsLanguage 'en', got '%s'", lyricsData.LyricsLanguage)
	}
}

// Test integration avec des données mock
func TestLyricsIntegration(t *testing.T) {
	client := New("test-api-key")

	// Test avec une requête qui devrait échouer (pas de serveur mock)
	_, err := client.Lyrics("", "Test Song")

	// On s'attend à une erreur de réseau puisqu'on ne mock pas HTTP
	if err == nil {
		t.Log("Warning: Expected network error, but got nil. This might indicate real API call.")
	}
}

// Benchmarks
func BenchmarkParseTimeString(b *testing.B) {
	timeStr := "03:45.67"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parseTimeString(timeStr)
	}
}

func BenchmarkParseQuery(b *testing.B) {
	query := "The Beatles - Hey Jude"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parseQuery(query)
	}
}

func BenchmarkParseLRCSubtitles(b *testing.B) {
	client := New("test-key")
	lrcContent := `[00:12.50]Line one
[00:17.20]Line two
[00:25.00]Line three
[00:32.10]Line four`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		client.parseLRCSubtitles(lrcContent)
	}
}

// Test helper pour créer des lyrics.Line de test
func createTestMusixMatchLyrics() []lyrics.Line {
	return []lyrics.Line{
		{Time: 12500, Words: "First line with timing"},
		{Time: 17200, Words: "Second line with timing"},
		{Time: 0, Words: "Plain text line without timing"},
	}
}

func TestCreateTestMusixMatchLyrics(t *testing.T) {
	lines := createTestMusixMatchLyrics()

	if len(lines) != 3 {
		t.Errorf("Expected 3 lines, got %d", len(lines))
	}

	// Premier et deuxième lignes ont du timing
	if lines[0].Time != 12500 {
		t.Errorf("Expected Time 12500 for first line, got %d", lines[0].Time)
	}

	if lines[1].Time != 17200 {
		t.Errorf("Expected Time 17200 for second line, got %d", lines[1].Time)
	}

	// Troisième ligne sans timing (lyrics réguliers)
	if lines[2].Time != 0 {
		t.Errorf("Expected Time 0 for third line, got %d", lines[2].Time)
	}
}
