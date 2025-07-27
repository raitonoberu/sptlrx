package lrclib

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/raitonoberu/sptlrx/lyrics"
)

const (
	BaseURL   = "https://lrclib.net/api"
	UserAgent = "sptlrx/1.2.3 (https://github.com/raitonoberu/sptlrx)"
)

var (
	ErrTrackNotFound = errors.New("track not found in LRCLib")
	ErrInvalidParams = errors.New("invalid search parameters")
)

// Client implements lyrics.Provider interface for LRCLib API
type Client struct {
	httpClient *http.Client
	baseURL    string
	userAgent  string
}

// LRCLibResponse represents the API response structure
type LRCLibResponse struct {
	ID           int    `json:"id"`
	TrackName    string `json:"trackName"`
	ArtistName   string `json:"artistName"`
	AlbumName    string `json:"albumName"`
	Duration     int    `json:"duration"`
	Instrumental bool   `json:"instrumental"`
	PlainLyrics  string `json:"plainLyrics"`
	SyncedLyrics string `json:"syncedLyrics"`
}

// ErrorResponse represents error responses from the API
type ErrorResponse struct {
	Code    int    `json:"code"`
	Name    string `json:"name"`
	Message string `json:"message"`
}

func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{Timeout: 10 * time.Second},
		baseURL:    BaseURL,
		userAgent:  UserAgent,
	}
}

// Lyrics implements the lyrics.Provider interface
func (c *Client) Lyrics(id, query string) ([]lyrics.Line, error) {
	// For LRCLib, we expect the query to contain track info
	// Format: "artist|track|album|duration"
	// This is a simplified approach - in V2 this would be more sophisticated

	// Parse query (simplified for now)
	trackInfo, err := parseQuery(query)
	if err != nil {
		return nil, fmt.Errorf("failed to parse query: %w", err)
	}

	// Get lyrics from LRCLib
	response, err := c.getLyrics(trackInfo)
	if err != nil {
		return nil, err
	}

	// Convert to sptlrx format
	return c.convertToLines(response)
}

// getLyrics fetches lyrics from LRCLib API
func (c *Client) getLyrics(info TrackInfo) (*LRCLibResponse, error) {
	// Build request URL
	reqURL := fmt.Sprintf("%s/get", c.baseURL)
	params := url.Values{}
	params.Add("artist_name", info.Artist)
	params.Add("track_name", info.Track)
	params.Add("album_name", info.Album)
	params.Add("duration", strconv.Itoa(info.Duration))

	fullURL := fmt.Sprintf("%s?%s", reqURL, params.Encode())

	// Create request
	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", c.userAgent)

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Handle response
	if resp.StatusCode == 404 {
		return nil, ErrTrackNotFound
	}

	if resp.StatusCode != 200 {
		var errorResp ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errorResp); err == nil {
			return nil, fmt.Errorf("API error %d: %s", errorResp.Code, errorResp.Message)
		}
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Parse response
	var lrcResponse LRCLibResponse
	if err := json.NewDecoder(resp.Body).Decode(&lrcResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &lrcResponse, nil
}

// convertToLines converts LRCLib response to sptlrx lyrics format
func (c *Client) convertToLines(response *LRCLibResponse) ([]lyrics.Line, error) {
	if response.Instrumental {
		return []lyrics.Line{{Time: 0, Words: "[Instrumental]"}}, nil
	}

	// Prefer synced lyrics if available
	lyricsText := response.SyncedLyrics
	if lyricsText == "" {
		lyricsText = response.PlainLyrics
	}

	if lyricsText == "" {
		return nil, ErrTrackNotFound
	}

	// Parse LRC format to lyrics.Line
	return parseLRCFormat(lyricsText), nil
}

// TrackInfo holds track information for API requests
type TrackInfo struct {
	Artist   string
	Track    string
	Album    string
	Duration int
}

// parseQuery parses the query string into TrackInfo
// TODO: In V2, this would be a more sophisticated metadata structure
func parseQuery(query string) (TrackInfo, error) {
	// Simplified parsing - in real implementation this would be more robust
	// For now, return empty to avoid compilation errors
	// This would be implemented based on how the player provides track info
	return TrackInfo{}, ErrInvalidParams
}

// parseLRCFormat converts LRC format string to lyrics.Line slice
func parseLRCFormat(lrcText string) []lyrics.Line {
	// TODO: Implement LRC parsing
	// This is a placeholder - real implementation would parse [mm:ss.xx] format
	lines := []lyrics.Line{
		{Time: 0, Words: "LRCLib integration in progress..."},
	}
	return lines
}
