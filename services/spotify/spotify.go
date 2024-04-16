package spotify

import (
	"strings"

	"sptlrx/lyrics"
	"sptlrx/player"

	lyricsapi "github.com/raitonoberu/lyricsapi/lyrics"
)

var ErrInvalidCookie = lyricsapi.ErrInvalidCookie

func New(cookie string) (*Client, error) {
	if cookie == "" {
		return nil, ErrInvalidCookie
	}
	return &Client{lyricsapi.NewLyricsApi(cookie)}, nil
}

// Client implements both player.Player and lyrics.Provider
type Client struct {
	api *lyricsapi.LyricsApi
}

func (c *Client) State() (*player.State, error) {
	result, err := c.api.State()
	if err != nil {
		return nil, err
	}
	if result == nil || result.Item == nil {
		return nil, nil
	}

	return &player.State{
		ID:       "spotify:" + result.Item.ID,
		Position: result.Progress,
		Playing:  result.Playing,
	}, nil
}

func (c *Client) Lyrics(id, query string) ([]lyrics.Line, error) {
	var (
		result *lyricsapi.LyricsResult
		err    error
	)
	if strings.HasPrefix(id, "spotify:") {
		result, err = c.api.GetByID(id[8:])
	} else {
		result, err = c.api.GetByName(query)
	}

	if err != nil {
		return nil, err
	}
	if result == nil || len(result.Lyrics.Lines) == 0 {
		return nil, nil
	}

	lines := make([]lyrics.Line, len(result.Lyrics.Lines))
	for i, l := range result.Lyrics.Lines {
		lines[i] = lyrics.Line(l)
	}
	return lines, nil
}
