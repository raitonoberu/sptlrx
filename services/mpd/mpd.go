package mpd

import (
	"sptlrx/player"
	"strconv"

	"github.com/fhs/gompd/mpd"
)

func New(address, password string) (*Client, error) {
	c := &Client{
		address:  address,
		password: password,
	}
	return c, c.connect()
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

	var query string
	if artist != "" {
		query = artist + " " + title
	} else {
		query = title
	}

	return &player.State{
		ID:       status["songid"],
		Query:    query,
		Playing:  status["state"] == "play",
		Position: int(elapsed) * 1000,
	}, nil
}
