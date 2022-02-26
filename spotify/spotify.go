package spotify

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"
)

var (
	ErrInvalidCookie = errors.New("invalid cookie provided")
)

const tokenUrl = "https://open.spotify.com/get_access_token?reason=transport&productType=web_player"
const lyricsUrl = "https://spclient.wg.spotify.com/color-lyrics/v2/track/"
const currentUrl = "https://api.spotify.com/v1/me/player/currently-playing"

func NewClient(cookie string) (*SpotifyClient, error) {
	c := &SpotifyClient{
		cookie: cookie,
	}
	return c, c.checkToken()
}

type SpotifyClient struct {
	cookie    string
	token     string
	expiresIn time.Time
}

func (c *SpotifyClient) Current() (*CurrentlyPlaying, error) {
	err := c.checkToken()
	if err != nil {
		return nil, err
	}
	req, _ := http.NewRequest("GET", currentUrl, nil)
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
	return &CurrentlyPlaying{
		ID:       result.Item.ID,
		Position: result.Progress,
		Playing:  result.Playing,
	}, nil
}

func (c *SpotifyClient) Lyrics(spotifyID string) (LyricsLines, error) {
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
	if result.Lyrics.Lines == nil {
		// not found
		return nil, nil
	}
	return result.Lyrics.Lines, nil
}

func (c *SpotifyClient) checkToken() error {
	if !c.tokenExpired() {
		return nil
	}
	return c.updateToken()
}

func (c *SpotifyClient) tokenExpired() bool {
	return c.token == "" || time.Now().After(c.expiresIn)
}

func (c *SpotifyClient) updateToken() error {
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
		Lines []*LyricsLine `json:"lines"`
	} `json:"lyrics"`
}

type currentBody struct {
	Progress int  `json:"progress_ms"`
	Playing  bool `json:"is_playing"`
	Item     *struct {
		ID string `json:"id"`
	} `json:"item"`
}
