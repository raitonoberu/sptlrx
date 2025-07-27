package musixmatch

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
	baseURL   = "https://api.musixmatch.com/ws/1.1"
	userAgent = "sptlrx v1.2.3 (https://github.com/raitonoberu/sptlrx)"
)

// Client implements lyrics.Provider for MusixMatch API
type Client struct {
	baseURL    string
	httpClient *http.Client
	userAgent  string
	apiKey     string
}

// New creates a new MusixMatch client
func New(apiKey string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
		userAgent: userAgent,
		apiKey:    apiKey,
	}
}

// MusixMatchResponse represents the common response structure
type MusixMatchResponse struct {
	Message struct {
		Header struct {
			StatusCode int    `json:"status_code"`
			Available  int    `json:"available"`
			Hint       string `json:"hint,omitempty"`
		} `json:"header"`
		Body interface{} `json:"body"`
	} `json:"message"`
}

// Track represents a MusixMatch track
type Track struct {
	TrackID            int    `json:"track_id"`
	TrackName          string `json:"track_name"`
	ArtistName         string `json:"artist_name"`
	AlbumName          string `json:"album_name"`
	TrackLength        int    `json:"track_length"`
	HasLyrics          int    `json:"has_lyrics"`
	HasSubtitles       int    `json:"has_subtitles"`
	HasLyricsVocals    int    `json:"has_lyrics_vocals"`
	HasLyricsVocalsBoP int    `json:"has_lyrics_vocals_bop"`
	TrackShareURL      string `json:"track_share_url"`
}

// TrackSearchResponse represents track search response
type TrackSearchResponse struct {
	TrackList []struct {
		Track Track `json:"track"`
	} `json:"track_list"`
}

// LyricsData represents lyrics from MusixMatch
type LyricsData struct {
	LyricsID        int    `json:"lyrics_id"`
	LyricsBody      string `json:"lyrics_body"`
	LyricsLanguage  string `json:"lyrics_language"`
	ScriptTracking  string `json:"script_tracking_url"`
	PixelTracking   string `json:"pixel_tracking_url"`
	LyricsCopyright string `json:"lyrics_copyright"`
	UpdatedTime     string `json:"updated_time"`
}

// LyricsResponse represents lyrics response
type LyricsResponse struct {
	Lyrics LyricsData `json:"lyrics"`
}

// SubtitleData represents subtitles (time-synced lyrics)
type SubtitleData struct {
	SubtitleID       int    `json:"subtitle_id"`
	SubtitleBody     string `json:"subtitle_body"`
	SubtitleLanguage string `json:"subtitle_language"`
	ScriptTracking   string `json:"script_tracking_url"`
	PixelTracking    string `json:"pixel_tracking_url"`
	UpdatedTime      string `json:"updated_time"`
}

// SubtitleResponse represents subtitle response
type SubtitleResponse struct {
	Subtitle SubtitleData `json:"subtitle"`
}

// SubtitleLine represents a single line in LRC format
type SubtitleLine struct {
	Time string `json:"time"`
	Text string `json:"text"`
}

// Lyrics implements lyrics.Provider interface
func (c *Client) Lyrics(id, query string) ([]lyrics.Line, error) {
	if c.apiKey == "" {
		return nil, fmt.Errorf("MusixMatch API key is required")
	}

	if query == "" {
		return nil, fmt.Errorf("query cannot be empty")
	}

	// Parse query to extract artist and track
	artist, track := parseQuery(query)

	// Search for the track
	tracks, err := c.searchTracks(artist, track)
	if err != nil {
		return nil, fmt.Errorf("failed to search tracks: %w", err)
	}

	if len(tracks) == 0 {
		return nil, nil // No tracks found
	}

	// Try to get subtitles (time-synced lyrics) first
	for _, trackInfo := range tracks {
		if trackInfo.HasSubtitles == 1 {
			lines, err := c.getSubtitles(trackInfo.TrackID)
			if err == nil && len(lines) > 0 {
				return lines, nil
			}
		}
	}

	// Fallback to regular lyrics (not time-synced)
	for _, trackInfo := range tracks {
		if trackInfo.HasLyrics == 1 {
			lines, err := c.getLyrics(trackInfo.TrackID)
			if err == nil && len(lines) > 0 {
				return lines, nil
			}
		}
	}

	return nil, nil // No lyrics found
}

// searchTracks searches for tracks on MusixMatch
func (c *Client) searchTracks(artist, track string) ([]Track, error) {
	params := url.Values{}
	params.Set("apikey", c.apiKey)
	params.Set("format", "json")
	params.Set("page_size", "10")

	if artist != "" && track != "" {
		params.Set("q_artist", artist)
		params.Set("q_track", track)
	} else {
		// If we can't parse artist/track, search in track name
		params.Set("q_track", track)
	}

	requestURL := fmt.Sprintf("%s/track.search?%s", c.baseURL, params.Encode())

	req, err := http.NewRequest("GET", requestURL, nil)
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
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var response MusixMatchResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if response.Message.Header.StatusCode != 200 {
		return nil, fmt.Errorf("API error: status %d", response.Message.Header.StatusCode)
	}

	// Parse the body as TrackSearchResponse
	bodyJSON, err := json.Marshal(response.Message.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal body: %w", err)
	}

	var searchResp TrackSearchResponse
	if err := json.Unmarshal(bodyJSON, &searchResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal search response: %w", err)
	}

	var tracks []Track
	for _, item := range searchResp.TrackList {
		tracks = append(tracks, item.Track)
	}

	return tracks, nil
}

// getLyrics gets regular lyrics for a track
func (c *Client) getLyrics(trackID int) ([]lyrics.Line, error) {
	params := url.Values{}
	params.Set("apikey", c.apiKey)
	params.Set("track_id", fmt.Sprintf("%d", trackID))
	params.Set("format", "json")

	requestURL := fmt.Sprintf("%s/track.lyrics.get?%s", c.baseURL, params.Encode())

	req, err := http.NewRequest("GET", requestURL, nil)
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
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var response MusixMatchResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if response.Message.Header.StatusCode != 200 {
		return nil, fmt.Errorf("API error: status %d", response.Message.Header.StatusCode)
	}

	// Parse the body as LyricsResponse
	bodyJSON, err := json.Marshal(response.Message.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal body: %w", err)
	}

	var lyricsResp LyricsResponse
	if err := json.Unmarshal(bodyJSON, &lyricsResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal lyrics response: %w", err)
	}

	if lyricsResp.Lyrics.LyricsBody == "" {
		return nil, nil
	}

	// Convert plain text lyrics to lyrics.Line format
	lines := strings.Split(lyricsResp.Lyrics.LyricsBody, "\n")
	var result []lyrics.Line

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.Contains(line, "******* This Lyrics is NOT for Commercial use *******") {
			result = append(result, lyrics.Line{
				Time:  0, // Regular lyrics don't have timing
				Words: line,
			})
		}
	}

	return result, nil
}

// getSubtitles gets time-synced subtitles for a track
func (c *Client) getSubtitles(trackID int) ([]lyrics.Line, error) {
	params := url.Values{}
	params.Set("apikey", c.apiKey)
	params.Set("track_id", fmt.Sprintf("%d", trackID))
	params.Set("format", "json")

	requestURL := fmt.Sprintf("%s/track.subtitle.get?%s", c.baseURL, params.Encode())

	req, err := http.NewRequest("GET", requestURL, nil)
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
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var response MusixMatchResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if response.Message.Header.StatusCode != 200 {
		return nil, fmt.Errorf("API error: status %d", response.Message.Header.StatusCode)
	}

	// Parse the body as SubtitleResponse
	bodyJSON, err := json.Marshal(response.Message.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal body: %w", err)
	}

	var subtitleResp SubtitleResponse
	if err := json.Unmarshal(bodyJSON, &subtitleResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal subtitle response: %w", err)
	}

	if subtitleResp.Subtitle.SubtitleBody == "" {
		return nil, nil
	}

	// Parse LRC format subtitles
	return c.parseLRCSubtitles(subtitleResp.Subtitle.SubtitleBody)
}

// parseLRCSubtitles parses LRC format subtitles from MusixMatch
func (c *Client) parseLRCSubtitles(lrcContent string) ([]lyrics.Line, error) {
	// MusixMatch uses a specific LRC format, delegate to existing LRC parser
	// For now, we'll implement a basic parser
	lines := strings.Split(lrcContent, "\n")
	var result []lyrics.Line

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Basic LRC parsing - look for [mm:ss.xx] format
		if strings.HasPrefix(line, "[") && strings.Contains(line, "]") {
			timeEnd := strings.Index(line, "]")
			if timeEnd > 0 {
				timeStr := line[1:timeEnd]
				text := strings.TrimSpace(line[timeEnd+1:])

				// Parse time (basic implementation)
				timeMs := parseTimeString(timeStr)
				if timeMs >= 0 && text != "" {
					result = append(result, lyrics.Line{
						Time:  timeMs,
						Words: text,
					})
				}
			}
		}
	}

	return result, nil
}

// parseTimeString parses time string in mm:ss.xx format to milliseconds
func parseTimeString(timeStr string) int {
	// Basic parsing for mm:ss.xx format
	parts := strings.Split(timeStr, ":")
	if len(parts) != 2 {
		return -1
	}

	// Parse minutes
	var minutes int
	fmt.Sscanf(parts[0], "%d", &minutes)

	// Parse seconds and milliseconds
	secParts := strings.Split(parts[1], ".")
	var seconds, milliseconds int
	fmt.Sscanf(secParts[0], "%d", &seconds)
	if len(secParts) > 1 {
		// Convert centiseconds to milliseconds
		fmt.Sscanf(secParts[1], "%d", &milliseconds)
		milliseconds *= 10
	}

	return minutes*60*1000 + seconds*1000 + milliseconds
}

// parseQuery attempts to extract artist and track from a query string
func parseQuery(query string) (artist, track string) {
	// Simple parsing: "Artist - Track"
	if strings.Contains(query, " - ") {
		parts := strings.SplitN(query, " - ", 2)
		if len(parts) == 2 {
			return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
		}
	}

	// Fallback: assume entire query is track name
	return "", query
}
