package hosted

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"github.com/raitonoberu/sptlrx/lyrics"
)

// Host your own: https://github.com/raitonoberu/lyricsapi
func New(host string) *Client {
	return &Client{
		host: host,
	}
}

// Client implements lyrics.Provider
type Client struct {
	host string
}

func (c *Client) Lyrics(id, query string) ([]lyrics.Line, error) {
	var url = fmt.Sprintf("https://%s/api/lyrics?name=%s", c.host, url.QueryEscape(query))

	req, _ := http.NewRequest("GET", url, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result []lyrics.Line
	err = json.NewDecoder(resp.Body).Decode(&result)
	return result, err
}
