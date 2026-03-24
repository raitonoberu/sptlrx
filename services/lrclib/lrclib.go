package lrclib

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"strings"

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
func (c *Client) Lyrics(id, query string) ([]lyrics.Line, error) {
	url := "https://lrclib.net/api/search?" + url.Values{
		"q": {query},
	}.Encode()
	req, err := http.NewRequest("GET", url, nil)
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
		return parceSynced(t)
	}
	if t.PlainLyrics != "" {
		return parsePlain(t)
	}
	return nil
}

func parceSynced(r lrclibTrack) []lyrics.Line {
	lines := strings.Split(r.SyncedLyrics, "\n")
	result := make([]lyrics.Line, len(lines))
	for i, line := range lines {
		result[i] = parseLrcLine(line)
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

func parseLrcLine(line string) lyrics.Line {
	// "[00:17.12] whatever"
	if len(line) < 11 {
		return lyrics.Line{}
	}
	m, _ := strconv.Atoi(line[1:3])
	s, _ := strconv.Atoi(line[4:6])
	ms, _ := strconv.Atoi(line[7:9])
	return lyrics.Line{
		Time:  m*60*1000 + s*1000 + ms*10,
		Words: line[11:],
	}
}
