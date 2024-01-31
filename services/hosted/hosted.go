package hosted

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
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
		os.Stderr.WriteString("HOSTD: Could not find lyrics" + "\n")
		return nil, err
	}
	defer resp.Body.Close()

	var result []lyrics.Line
	err = json.NewDecoder(resp.Body).Decode(&result)
	if len(result) > 0 {
		// var header []lyrics.Line
		// header = append(header, lyrics.Line{
		// 	Time:  0,
		// 	Words: "Loading from Hosted...",
		// })
		// if result[0].Time < 10 {
		// 	result[0].Time = 10
		// }
		os.Stderr.WriteString("HOSTD: Found Lyrics" + "\n")
		return result, nil
	} else {
		os.Stderr.WriteString("HOSTD: Empty Lyrics" + "\n")
		return nil, err
	}
}

func (c *Client) Name() string {
	return "HOSTD"
}
