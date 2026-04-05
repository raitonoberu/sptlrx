package lrclib

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
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
	if artist != "" && track != "" {
		return c.get(artist, track)
	}
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
		if !lyrics.IsTimestampLine(line) {
			continue
		}
		result = append(result, lyrics.ParseLrcLine(line))
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
