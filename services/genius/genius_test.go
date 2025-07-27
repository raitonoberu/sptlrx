package genius

import (
	"testing"

	"github.com/raitonoberu/sptlrx/lyrics"
)

// MockHTTPClient pour les tests
type MockHTTPResponse struct {
	statusCode int
	body       string
}

func TestNew(t *testing.T) {
	client := New("test-token")

	if client == nil {
		t.Fatal("New should return a non-nil client")
	}

	if client.accessToken != "test-token" {
		t.Errorf("Expected access token 'test-token', got '%s'", client.accessToken)
	}

	if client.baseURL != baseURL {
		t.Errorf("Expected baseURL '%s', got '%s'", baseURL, client.baseURL)
	}
}

func TestNewWithoutToken(t *testing.T) {
	client := New("")

	if client == nil {
		t.Fatal("New should return a non-nil client even without token")
	}

	if client.accessToken != "" {
		t.Errorf("Expected empty access token, got '%s'", client.accessToken)
	}
}

func TestLyricsEmptyQuery(t *testing.T) {
	client := New("")

	_, err := client.Lyrics("", "")
	if err == nil {
		t.Error("Expected error for empty query")
	}

	if err.Error() != "query cannot be empty" {
		t.Errorf("Expected 'query cannot be empty', got '%s'", err.Error())
	}
}

func TestCleanHTML(t *testing.T) {
	client := New("")

	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Basic HTML tags",
			input:    "<p>Hello <b>world</b></p>",
			expected: "Hello \nworld",
		},
		{
			name:     "HTML entities",
			input:    "Rock &amp; roll &lt;music&gt;",
			expected: "Rock & roll <music>",
		},
		{
			name:     "Multiple newlines",
			input:    "Line 1\n\n\nLine 2",
			expected: "Line 1\nLine 2",
		},
		{
			name:     "Multiple spaces",
			input:    "Word1     Word2",
			expected: "Word1 Word2",
		},
		{
			name:     "Mixed content",
			input:    "<div>Hello &amp; <span>goodbye</span></div>\n\n<p>Next line</p>",
			expected: "Hello & \ngoodbye\n\nNext line",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := client.cleanHTML(tc.input)
			if result != tc.expected {
				t.Errorf("Expected '%s', got '%s'", tc.expected, result)
			}
		})
	}
}

func TestExtractLyricsFromHTML(t *testing.T) {
	client := New("")

	testCases := []struct {
		name     string
		html     string
		expected string
	}{
		{
			name: "Data lyrics container",
			html: `<div data-lyrics-container="true">
				<p>Verse 1</p>
				<p>Chorus line</p>
			</div>`,
			expected: "Verse 1\n\nChorus line",
		},
		{
			name: "Lyrics class",
			html: `<div class="lyrics">
				Line 1<br/>
				Line 2
			</div>`,
			expected: "Line 1\n\nLine 2",
		},
		{
			name:     "No lyrics found",
			html:     `<div class="other-content">Not lyrics</div>`,
			expected: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := client.extractLyricsFromHTML(tc.html)
			if result != tc.expected {
				t.Errorf("Expected '%s', got '%s'", tc.expected, result)
			}
		})
	}
}

// Test integration avec des données mock
func TestLyricsIntegration(t *testing.T) {
	// Ce test nécessiterait un mock HTTP client
	// Pour l'instant, on teste la structure de base

	client := New("test-token")

	// Test avec une requête qui devrait échouer (pas de serveur mock)
	_, err := client.Lyrics("", "Test Song")

	// On s'attend à une erreur de réseau puisqu'on ne mock pas HTTP
	if err == nil {
		t.Log("Warning: Expected network error, but got nil. This might indicate real API call.")
	}
}

func TestGeniusSongStruct(t *testing.T) {
	song := GeniusSong{
		ID:          12345,
		Title:       "Test Song",
		FullTitle:   "Test Song by Test Artist",
		URL:         "https://genius.com/test",
		LyricsState: "complete",
		Artist:      "Test Artist",
	}

	if song.ID != 12345 {
		t.Errorf("Expected ID 12345, got %d", song.ID)
	}

	if song.Title != "Test Song" {
		t.Errorf("Expected title 'Test Song', got '%s'", song.Title)
	}

	if song.LyricsState != "complete" {
		t.Errorf("Expected lyrics state 'complete', got '%s'", song.LyricsState)
	}
}

// Benchmark pour les fonctions de nettoyage HTML
func BenchmarkCleanHTML(b *testing.B) {
	client := New("")
	html := "<div>Hello <b>world</b> &amp; <span>universe</span></div>\n\n<p>Next line</p>"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		client.cleanHTML(html)
	}
}

func BenchmarkExtractLyricsFromHTML(b *testing.B) {
	client := New("")
	html := `<div data-lyrics-container="true">
		<p>Line 1</p>
		<p>Line 2</p>
		<p>Line 3</p>
	</div>`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		client.extractLyricsFromHTML(html)
	}
}

// Test helper pour créer des lyrics.Line de test
func createTestLyrics() []lyrics.Line {
	return []lyrics.Line{
		{Time: 0, Words: "First line"},
		{Time: 0, Words: "Second line"},
		{Time: 0, Words: "Third line"},
	}
}

func TestCreateTestLyrics(t *testing.T) {
	lines := createTestLyrics()

	if len(lines) != 3 {
		t.Errorf("Expected 3 lines, got %d", len(lines))
	}

	if lines[0].Words != "First line" {
		t.Errorf("Expected 'First line', got '%s'", lines[0].Words)
	}

	// Genius ne fournit pas de timing, donc toutes les lignes ont Time: 0
	for i, line := range lines {
		if line.Time != 0 {
			t.Errorf("Expected Time 0 for line %d, got %d", i, line.Time)
		}
	}
}
