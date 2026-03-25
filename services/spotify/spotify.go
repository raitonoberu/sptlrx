package spotify

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/raitonoberu/sptlrx/player"
	"github.com/raitonoberu/sptlrx/services/spotify/auth"
)

func New() (*Client, error) {
	auth, err := auth.Load()
	if errors.Is(err, os.ErrNotExist) {
		err = errors.New("you must run `sptlrx login` first to use Spotify as a player")
	}
	if err != nil {
		return nil, err
	}

	return &Client{
		auth: auth,
	}, nil
}

// Client implements player.Player
type Client struct {
	auth *auth.Auth
	http http.Client
}

func (c *Client) State() (*player.State, error) {
	token, err := c.auth.GetToken(context.Background())
	if err != nil {
		return nil, err
	}

	req, _ := http.NewRequest("GET", "https://api.spotify.com/v1/me/player/currently-playing", nil)
	req.Header.Add("Authorization", "Bearer "+token)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNoContent {
		return nil, nil
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("status code %d", resp.StatusCode)
	}

	var state state
	if err := json.NewDecoder(resp.Body).Decode(&state); err != nil {
		return nil, err
	}

	query := ""
	for _, a := range state.Item.Artists {
		query += a.Name + " "
	}
	query += state.Item.Name

	return &player.State{
		ID:       state.Item.ID,
		Query:    query,
		Position: state.ProgressMs,
		Playing:  state.IsPlaying,
	}, nil
}

type state struct {
	IsPlaying  bool  `json:"is_playing"`
	ProgressMs int   `json:"progress_ms"`
	Item       track `json:"item"`
}

type track struct {
	ID      string        `json:"id"`
	Name    string        `json:"name"`
	Artists []trackArtist `json:"artists"`
}

type trackArtist struct {
	Name string `json:"name"`
}
