package genius

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/raitonoberu/sptlrx/lyrics"
)

const (
	baseURL   = "https://api.genius.com"
	userAgent = "sptlrx v1.2.3 (https://github.com/raitonoberu/sptlrx)"
)

// Client implements lyrics.Provider for Genius API
type Client struct {
	baseURL     string
	httpClient  *http.Client
	userAgent   string
	accessToken string
}

// New creates a new Genius client
// Note: Genius API requires an access token for full functionality
// For public search, we can use limited access without token
func New(accessToken string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
		userAgent:   userAgent,
		accessToken: accessToken,
	}
}

// GeniusSearchResponse represents the response from Genius search API
type GeniusSearchResponse struct {
	Meta struct {
		Status int `json:"status"`
	} `json:"meta"`
	Response struct {
		Hits []struct {
			Type   string `json:"type"`
			Result struct {
				ID                int    `json:"id"`
				Title             string `json:"title"`
				TitleWithFeatured string `json:"title_with_featured"`
				FullTitle         string `json:"full_title"`
				URL               string `json:"url"`
				LyricsState       string `json:"lyrics_state"`
				PrimaryArtist     struct {
					ID   int    `json:"id"`
					Name string `json:"name"`
				} `json:"primary_artist"`
			} `json:"result"`
		} `json:"hits"`
	} `json:"response"`
}

// GeniusSong represents a song with lyrics information
type GeniusSong struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	FullTitle   string `json:"full_title"`
	URL         string `json:"url"`
	LyricsState string `json:"lyrics_state"`
	Artist      string `json:"artist"`
}

// Lyrics implements lyrics.Provider interface
func (c *Client) Lyrics(id, query string) ([]lyrics.Line, error) {
	if query == "" {
		return nil, fmt.Errorf("query cannot be empty")
	}

	// Search for the song on Genius
	songs, err := c.searchSongs(query)
	if err != nil {
		return nil, fmt.Errorf("failed to search songs: %w", err)
	}

	if len(songs) == 0 {
		return nil, nil // No songs found
	}

	// Try to get lyrics from the best match
	for _, song := range songs {
		if song.LyricsState != "complete" {
			continue // Skip songs without complete lyrics
		}

		lyrics, err := c.scrapeLyrics(song.URL)
		if err != nil {
			continue // Try next song if scraping fails
		}

		if len(lyrics) > 0 {
			return lyrics, nil
		}
	}

	return nil, nil // No lyrics found
}

// searchSongs searches for songs on Genius API
func (c *Client) searchSongs(query string) ([]GeniusSong, error) {
	params := url.Values{}
	params.Set("q", query)

	requestURL := fmt.Sprintf("%s/search?%s", c.baseURL, params.Encode())

	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", c.userAgent)
	if c.accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.accessToken)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var searchResp GeniusSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if searchResp.Meta.Status != 200 {
		return nil, fmt.Errorf("API error: status %d", searchResp.Meta.Status)
	}

	// Convert hits to GeniusSong slice
	var songs []GeniusSong
	for _, hit := range searchResp.Response.Hits {
		if hit.Type == "song" {
			songs = append(songs, GeniusSong{
				ID:          hit.Result.ID,
				Title:       hit.Result.Title,
				FullTitle:   hit.Result.FullTitle,
				URL:         hit.Result.URL,
				LyricsState: hit.Result.LyricsState,
				Artist:      hit.Result.PrimaryArtist.Name,
			})
		}
	}

	return songs, nil
}

// scrapeLyrics scrapes lyrics from a Genius song page
// Note: This is a simplified scraper. In production, you might want to use
// a more robust HTML parser like goquery
func (c *Client) scrapeLyrics(songURL string) ([]lyrics.Line, error) {
	req, err := http.NewRequest("GET", songURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", c.userAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("page returned status %d", resp.StatusCode)
	}

	// Read the HTML content
	bodyBytes := make([]byte, 1024*1024) // 1MB buffer
	n, err := resp.Body.Read(bodyBytes)
	if err != nil && n == 0 {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	html := string(bodyBytes[:n])

	// Extract lyrics using regex (simplified approach)
	// Genius stores lyrics in various formats, this is a basic extraction
	lyricsText := c.extractLyricsFromHTML(html)
	if lyricsText == "" {
		return nil, fmt.Errorf("no lyrics found in page")
	}

	// Convert plain text to lyrics.Line format
	// Genius doesn't provide time-synced lyrics, so we'll use 0 time
	lines := strings.Split(lyricsText, "\n")
	var result []lyrics.Line

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			result = append(result, lyrics.Line{
				Time:  0,
				Words: line,
			})
		}
	}

	return result, nil
}

// extractLyricsFromHTML extracts lyrics text from Genius HTML page
func (c *Client) extractLyricsFromHTML(html string) string {
	// This is a simplified extraction. Genius uses different HTML structures
	// for different pages, so this might need adjustment

	// Try to find lyrics in common Genius HTML patterns
	patterns := []string{
		`<div[^>]*data-lyrics-container[^>]*>(.*?)</div>`,
		`<div[^>]*class="[^"]*lyrics[^"]*"[^>]*>(.*?)</div>`,
		`<p[^>]*>(.*?)</p>`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(`(?s)` + pattern) // (?s) makes . match newlines too
		matches := re.FindStringSubmatch(html)
		if len(matches) > 1 {
			// Clean HTML tags and decode entities
			lyrics := c.cleanHTML(matches[1])
			if lyrics != "" {
				return lyrics
			}
		}
	}

	return ""
}

// cleanHTML removes HTML tags and decodes basic entities
func (c *Client) cleanHTML(html string) string {
	// Remove HTML tags but preserve structure with newlines
	re := regexp.MustCompile(`<[^>]*>`)
	text := re.ReplaceAllString(html, "\n")

	// Decode basic HTML entities
	text = strings.ReplaceAll(text, "&amp;", "&")
	text = strings.ReplaceAll(text, "&lt;", "<")
	text = strings.ReplaceAll(text, "&gt;", ">")
	text = strings.ReplaceAll(text, "&quot;", "\"")
	text = strings.ReplaceAll(text, "&#x27;", "'")
	text = strings.ReplaceAll(text, "&nbsp;", " ")

	// Clean excessive whitespace but preserve intentional line breaks
	lines := strings.Split(text, "\n")
	var cleanLines []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		cleanLines = append(cleanLines, line)
	}

	text = strings.Join(cleanLines, "\n")

	// Clean multiple consecutive newlines to double newlines max
	re = regexp.MustCompile(`\n{3,}`)
	text = re.ReplaceAllString(text, "\n\n")

	return strings.TrimSpace(text)
} // GetSongInfo returns information about a song without lyrics
func (c *Client) GetSongInfo(query string) ([]GeniusSong, error) {
	return c.searchSongs(query)
}
