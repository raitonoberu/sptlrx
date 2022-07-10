package hosted

import (
	"encoding/json"
	"net/http"
	"net/url"
	"sptlrx/lyrics"
)

// Host your own: https://github.com/raitonoberu/lyricsapi
const URL = "https://lyricsapi.vercel.app/api/lyrics?"

func New() *Client {
	return &Client{}
}

// Client implements lyrics.Provider
type Client struct{}

func (c *Client) Lyrics(id, query string) ([]lyrics.Line, error) {
	var url = URL + url.Values{
		"name": {query},
	}.Encode()
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
