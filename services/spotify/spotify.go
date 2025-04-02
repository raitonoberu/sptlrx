package spotify

import (
	"strings"

	"github.com/raitonoberu/lyricsapi/spotify"
	"github.com/raitonoberu/sptlrx/lyrics"
	"github.com/raitonoberu/sptlrx/player"
)

var ErrInvalidCookie = spotify.ErrInvalidCookie

func New(cookie string) (*Client, error) {
	if cookie == "" {
		return nil, ErrInvalidCookie
	}
	return &Client{spotify.NewClient(cookie)}, nil
}

// Client implements both player.Player and lyrics.Provider
type Client struct {
	api *spotify.Client
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
		result []spotify.LyricsLine
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
	if len(result) == 0 {
		return nil, nil
	}

	lines := make([]lyrics.Line, len(result))
	for i, l := range result {
		lines[i] = lyrics.Line(l)
	}
	return lines, nil
}
