package lrclib

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/raitonoberu/sptlrx/lyrics"
)

const userAgent = "sptlrx v1.0.0 (https://github.com/raitonoberu/sptlrx)"

func New() *Client {
	return &Client{}
}

type Client struct {
	http http.Client
}

// Client implements lyrics.Provider
func (c *Client) Lyrics(artist, track string) ([]lyrics.Line, error) {
	// try /api/get first (exact match, no caching issues)
	lines, err := c.get(artist, track)
	if err == nil && lines != nil {
		return lines, nil
	}

	// fallback to /api/search
	return c.search(artist + " " + track)
}

func (c *Client) get(artist, track string) ([]lyrics.Line, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	u := "https://lrclib.net/api/get?" + url.Values{
		"artist_name": {artist},
		"track_name":  {track},
	}.Encode()
	req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	var response lrclibTrack
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return nil, err
	}

	return parseTrack(response), nil
}

func (c *Client) search(query string) ([]lyrics.Line, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	u := "https://lrclib.net/api/search?" + url.Values{
		"q": {query},
	}.Encode()
	req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var response []lrclibTrack
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return nil, err
	}

	if len(response) == 0 {
		return nil, nil
	}
	return parseTrack(response[0]), nil
}

type lrclibTrack struct {
	PlainLyrics  string `json:"plainLyrics"`
	SyncedLyrics string `json:"syncedLyrics"`
}

func parseTrack(t lrclibTrack) []lyrics.Line {
	if t.SyncedLyrics != "" {
		return parseSynced(t)
	}
	if t.PlainLyrics != "" {
		return parsePlain(t)
	}
	return nil
}

func parseSynced(r lrclibTrack) []lyrics.Line {
	lines := strings.Split(r.SyncedLyrics, "\n")
	result := make([]lyrics.Line, 0, len(lines))
	for _, line := range lines {
		if !isTimestampLine(line) {
			continue
		}
		result = append(result, parseLrcLine(line))
	}
	return result
}

func parsePlain(r lrclibTrack) []lyrics.Line {
	lines := strings.Split(r.PlainLyrics, "\n")
	result := make([]lyrics.Line, len(lines))
	for i, line := range lines {
		result[i] = lyrics.Line{Words: line}
	}
	return result
}

// isTimestampLine checks if a line starts with a timestamp like [00:17.12]
func isTimestampLine(line string) bool {
	if len(line) < 10 {
		return false
	}
	return line[0] == '[' &&
		line[3] == ':' &&
		line[6] == '.' &&
		line[1] >= '0' && line[1] <= '9' &&
		line[2] >= '0' && line[2] <= '9'
}

func parseLrcLine(line string) lyrics.Line {
	// "[00:17.12] text" or "[00:17.123]text"
	m, _ := strconv.Atoi(line[1:3])
	s, _ := strconv.Atoi(line[4:6])

	closeBracket := strings.IndexByte(line, ']')

	msStr := line[7:closeBracket]
	ms, _ := strconv.Atoi(msStr)
	if len(msStr) == 2 {
		ms *= 10
	} else if len(msStr) == 1 {
		ms *= 100
	}

	words := line[closeBracket+1:]
	if strings.HasPrefix(words, " ") {
		words = words[1:]
	}

	return lyrics.Line{
		Time:  m*60*1000 + s*1000 + ms,
		Words: words,
	}
}
