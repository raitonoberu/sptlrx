package hosted

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sptlrx/lyrics"
	"sptlrx/player"
)

// Host your own: https://github.com/raitonoberu/lyricsapi
func New(host string) lyrics.Provider {
	return &Client{
		host: host,
	}
}

// Client implements lyrics.Provider
type Client struct {
	host string
}

func (c *Client) Lyrics(state player.State) ([]lyrics.Line, error) {

	query := state.Artist + " " + state.Title
	var url = fmt.Sprintf("https://%s/api/lyrics?name=%s", c.host, url.QueryEscape(query))

	req, _ := http.NewRequest("GET", url, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result []lyrics.Line
	err = json.NewDecoder(resp.Body).Decode(&result)
	if len(result) > 0 {
		return result, err
	} else {
		return nil, err
	}
}

func (c *Client) Name() string {
	return "HOSTD"
}
