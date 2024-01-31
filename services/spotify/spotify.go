package spotify

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"os"
	"sptlrx/lyrics"
	"sptlrx/player"
	"strings"
	"time"
)

var (
	ErrInvalidCookie = errors.New("invalid or empty cookie provided")
)

const tokenUrl = "https://open.spotify.com/get_access_token?reason=transport&productType=web_player"
const lyricsUrl = "https://spclient.wg.spotify.com/color-lyrics/v2/track/"
const stateUrl = "https://api.spotify.com/v1/me/player/currently-playing"
const searchUrl = "https://api.spotify.com/v1/search?"

func NewProvider(cookie string) (lyrics.Provider, error) {
	if cookie == "" {
		return nil, ErrInvalidCookie
	}
	return &Client{cookie: cookie}, nil
}
func NewPlayer(cookie string) (player.Player, error) {
	if cookie == "" {
		return nil, ErrInvalidCookie
	}
	return &Client{cookie: cookie}, nil
}

// Client implements both player.Player and lyrics.Provider
type Client struct {
	cookie    string
	token     string
	expiresIn time.Time
}

func (c *Client) State() (*player.State, error) {
	err := c.checkToken()
	if err != nil {
		return nil, err
	}

	req, _ := http.NewRequest("GET", stateUrl, nil)
	req.Header = http.Header{
		"referer":          {"https://open.spotify.com/"},
		"origin":           {"https://open.spotify.com/"},
		"accept":           {"application/json"},
		"accept-language":  {"en"},
		"app-platform":     {"WebPlayer"},
		"sec-ch-ua-mobile": {"?0"},

		"Authorization": {"Bearer " + c.token},
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	result := &currentBody{}
	err = json.NewDecoder(resp.Body).Decode(result)
	if err != nil {
		if err == io.EOF {
			// stopped
			return nil, nil
		}
		return nil, err
	}
	if result.Item == nil {
		return nil, nil
	}
	return &player.State{
		ID:       "spotify:" + result.Item.ID,
		Position: result.Progress,
		Playing:  result.Playing,
	}, nil
}

func (c *Client) Lyrics(state player.State) ([]lyrics.Line, error) {
	if strings.HasPrefix(state.ID, "spotify:") {
		os.Stderr.WriteString("SPTFY: Found Lyrics" + "\n")
		return c.lyrics(state.ID[8:])
	}
	id, err := c.search(state.Artist + " " + state.Title)
	if err != nil {
		os.Stderr.WriteString("SPTFY: Error Finding Lyrics" + "\n")
		return nil, err
	}
	lys, err := c.lyrics(id)
	if len(lys) > 0 && err != nil {
		// var header []lyrics.Line
		// header = append(header, lyrics.Line{
		// 	Time:  0,
		// 	Words: "Loading from Spotify...",
		// })
		// if lys[0].Time < 10 {
		// 	lys[0].Time = 10
		// }
		os.Stderr.WriteString("SPTFY: Found Lyrics" + "\n")
		return lys, err
	} else {
		os.Stderr.WriteString("SPTFY: Empty Lyrics" + "\n")
		return nil, err
	}
}

func (c *Client) search(query string) (string, error) {
	err := c.checkToken()
	if err != nil {
		return "", err
	}

	url := searchUrl + url.Values{
		"limit": {"1"},
		"type":  {"track"},
		"q":     {query},
	}.Encode()
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+c.token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	result := &searchBody{}
	err = json.NewDecoder(resp.Body).Decode(result)
	if err != nil {
		return "", err
	}
	if result.Tracks.Total == 0 {
		return "", nil
	}
	return result.Tracks.Items[0].ID, nil
}

func (c *Client) lyrics(spotifyID string) ([]lyrics.Line, error) {
	err := c.checkToken()
	if err != nil {
		return nil, err
	}

	req, _ := http.NewRequest("GET", lyricsUrl+spotifyID, nil)
	req.Header = http.Header{
		"referer":          {"https://open.spotify.com/"},
		"origin":           {"https://open.spotify.com/"},
		"accept":           {"application/json"},
		"accept-language":  {"en"},
		"app-platform":     {"WebPlayer"},
		"sec-ch-ua-mobile": {"?0"},

		"Authorization": {"Bearer " + c.token},
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	result := &lyricsBody{}
	err = json.NewDecoder(resp.Body).Decode(result)
	if err != nil {
		if err == io.EOF {
			// no lyrics
			return nil, nil
		}
		return nil, err
	}

	lines := make([]lyrics.Line, len(result.Lyrics.Lines))
	for i, l := range result.Lyrics.Lines {
		lines[i] = lyrics.Line(l)
	}

	return lines, nil
}

func (c *Client) checkToken() error {
	if !c.tokenExpired() {
		return nil
	}
	return c.updateToken()
}

func (c *Client) tokenExpired() bool {
	return c.token == "" || time.Now().After(c.expiresIn)
}

func (c *Client) updateToken() error {
	req, _ := http.NewRequest("GET", tokenUrl, nil)
	req.Header = http.Header{
		"referer":             {"https://open.spotify.com/"},
		"origin":              {"https://open.spotify.com/"},
		"accept":              {"application/json"},
		"accept-language":     {"en"},
		"app-platform":        {"WebPlayer"},
		"sec-fetch-dest":      {"empty"},
		"sec-fetch-mode":      {"cors"},
		"sec-fetch-site":      {"same-origin"},
		"spotify-app-version": {"1.1.54.35.ge9dace1d"},
		"cookie":              {c.cookie},
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	result := &tokenBody{}
	err = json.NewDecoder(resp.Body).Decode(result)
	if err != nil {
		return err
	}

	if result.IsAnonymous {
		return ErrInvalidCookie
	}

	if result.AccessToken == "" {
		return errors.New("couldn't get access token")
	}

	c.token = result.AccessToken
	c.expiresIn = time.Unix(0, result.ExpiresIn*int64(time.Millisecond))

	return nil
}

type tokenBody struct {
	AccessToken string `json:"accessToken"`
	ExpiresIn   int64  `json:"accessTokenExpirationTimestampMs"`
	IsAnonymous bool   `json:"isAnonymous"`
}

type lyricsBody struct {
	Lyrics struct {
		Lines []struct {
			Time  int    `json:"startTimeMs,string"`
			Words string `json:"words"`
		} `json:"lines"`
	} `json:"lyrics"`
}

type currentBody struct {
	Progress int  `json:"progress_ms"`
	Playing  bool `json:"is_playing"`
	Item     *struct {
		ID string `json:"id"`
	} `json:"item"`
}

type searchBody struct {
	Tracks struct {
		Items []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"items"`
		Total int `json:"total"`
	} `json:"tracks"`
}

func (c *Client) Name() string {
	return "SPOTI"
}
