package mpd

import (
	"strconv"

	"github.com/raitonoberu/sptlrx/player"

	"github.com/fhs/gompd/mpd"
)

func New(address, password string) (*Client, error) {
	c := &Client{
		address:  address,
		password: password,
	}
	// Attempt initial connection so callers can decide to skip on error
	if err := c.connect(); err != nil {
		return nil, err
	}
	return c, nil
}

// Client implements player.Player
type Client struct {
	address  string
	password string
	client   *mpd.Client
}

func (c *Client) connect() error {
	if c.client != nil {
		c.client.Close()
	}
	client, err := mpd.DialAuthenticated("tcp", c.address, c.password)
	if err != nil {
		c.client = nil
		return err
	}
	c.client = client
	return nil
}

func (c *Client) checkConnection() error {
	if c.client == nil || c.client.Ping() != nil {
		return c.connect()
	}
	return nil
}

func (c *Client) State() (*player.State, error) {
	if err := c.checkConnection(); err != nil {
		return nil, err
	}

	status, err := c.client.Status()
	if err != nil {
		return nil, err
	}
	current, err := c.client.CurrentSong()
	if err != nil {
		return nil, err
	}
	elapsed, _ := strconv.ParseFloat(status["elapsed"], 32)

	var title string
	if t, ok := current["Title"]; ok {
		title = t
	}

	var artist string
	if a, ok := current["Artist"]; ok {
		artist = a
	}

	return &player.State{
		ID:       status["songid"],
		Artist:   artist,
		Track:    title,
		Playing:  status["state"] == "play",
		Position: int(elapsed * 1000), // secs to ms
	}, nil
}
