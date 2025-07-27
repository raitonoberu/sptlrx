package lrclib

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/raitonoberu/sptlrx/lyrics"
)

const (
	baseURL   = "https://lrclib.net"
	userAgent = "sptlrx v1.2.3 (https://github.com/raitonoberu/sptlrx)"
)

// Client implements lyrics.Provider for LRCLib API
type Client struct {
	baseURL    string
	httpClient *http.Client
	userAgent  string
}

// New creates a new LRCLib client
func New() *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		userAgent: userAgent,
	}
}

// LRCLibResponse represents the response from LRCLib API
type LRCLibResponse struct {
	ID           int     `json:"id"`
	TrackName    string  `json:"trackName"`
	ArtistName   string  `json:"artistName"`
	AlbumName    string  `json:"albumName"`
	Duration     float64 `json:"duration"` // API returns duration as float
	Instrumental bool    `json:"instrumental"`
	PlainLyrics  string  `json:"plainLyrics"`
	SyncedLyrics string  `json:"syncedLyrics"`
}

// SearchStrategy defines the search approach
type SearchStrategy int

const (
	ExactMatch  SearchStrategy = iota // /api/get
	CachedOnly                        // /api/get-cached
	FuzzySearch                       // /api/search
)

// Lyrics implements lyrics.Provider interface
func (c *Client) Lyrics(id, query string) ([]lyrics.Line, error) {
	if query == "" {
		return nil, fmt.Errorf("query cannot be empty")
	}

	// Parse query to extract artist and track
	artist, track := parseQuery(query)
	if artist == "" || track == "" {
		return c.searchFuzzy(query)
	}

	// Try exact match first
	if lines, err := c.getLyrics(artist, track, "", 0, ExactMatch); err == nil && len(lines) > 0 {
		return lines, nil
	}

	// Fallback to cached search
	if lines, err := c.getLyrics(artist, track, "", 0, CachedOnly); err == nil && len(lines) > 0 {
		return lines, nil
	}

	// Final fallback to fuzzy search
	return c.searchFuzzy(query)
}

// getLyrics retrieves lyrics using the specified strategy
func (c *Client) getLyrics(artist, track, album string, duration float64, strategy SearchStrategy) ([]lyrics.Line, error) {
	var endpoint string
	params := url.Values{}

	switch strategy {
	case ExactMatch:
		endpoint = "/api/get"
		params.Set("artist_name", artist)
		params.Set("track_name", track)
		if album != "" {
			params.Set("album_name", album)
		}
		if duration > 0 {
			params.Set("duration", fmt.Sprintf("%.0f", duration))
		}
	case CachedOnly:
		endpoint = "/api/get-cached"
		params.Set("artist_name", artist)
		params.Set("track_name", track)
		if album != "" {
			params.Set("album_name", album)
		}
		if duration > 0 {
			params.Set("duration", fmt.Sprintf("%.0f", duration))
		}
	default:
		return nil, fmt.Errorf("unsupported strategy for getLyrics")
	}

	url := fmt.Sprintf("%s%s?%s", c.baseURL, endpoint, params.Encode())

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", c.userAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return nil, nil // No lyrics found
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var lrcResp LRCLibResponse
	if err := json.NewDecoder(resp.Body).Decode(&lrcResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if lrcResp.Instrumental {
		return nil, nil // Track is instrumental
	}

	// Prefer synced lyrics, fallback to plain lyrics
	if lrcResp.SyncedLyrics != "" {
		return ParseLRC(lrcResp.SyncedLyrics)
	}

	if lrcResp.PlainLyrics != "" {
		return []lyrics.Line{{Time: 0, Words: lrcResp.PlainLyrics}}, nil
	}

	return nil, nil
}

// searchFuzzy performs fuzzy search using /api/search
func (c *Client) searchFuzzy(query string) ([]lyrics.Line, error) {
	params := url.Values{}
	params.Set("q", query)

	url := fmt.Sprintf("%s/api/search?%s", c.baseURL, params.Encode())

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create search request: %w", err)
	}

	req.Header.Set("User-Agent", c.userAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("search request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("search API returned status %d", resp.StatusCode)
	}

	var results []LRCLibResponse
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, fmt.Errorf("failed to decode search response: %w", err)
	}

	// Take the first non-instrumental result
	for _, result := range results {
		if result.Instrumental {
			continue
		}

		if result.SyncedLyrics != "" {
			return ParseLRC(result.SyncedLyrics)
		}

		if result.PlainLyrics != "" {
			return []lyrics.Line{{Time: 0, Words: result.PlainLyrics}}, nil
		}
	}

	return nil, nil
}

// parseQuery attempts to extract artist and track from a query string
func parseQuery(query string) (artist, track string) {
	// Simple parsing: "Artist - Track" or "Artist Track"
	if strings.Contains(query, " - ") {
		parts := strings.SplitN(query, " - ", 2)
		if len(parts) == 2 {
			return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
		}
	}

	// Fallback: split on first space and assume first word is artist
	parts := strings.Fields(query)
	if len(parts) >= 2 {
		return parts[0], strings.Join(parts[1:], " ")
	}

	return "", query
}
