package spotify

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

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
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	token, err := c.auth.GetToken(ctx)
	if err != nil {
		return nil, err
	}

	req, _ := http.NewRequestWithContext(ctx, "GET", "https://api.spotify.com/v1/me/player/currently-playing", nil)
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

	artist := ""
	for i, a := range state.Item.Artists {
		if i != 0 {
			artist += " "
		}
		artist += a.Name
	}

	return &player.State{
		ID:       state.Item.ID,
		Artist:   artist,
		Track:    state.Item.Name,
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
