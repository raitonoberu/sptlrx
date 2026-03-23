package spotify

import (
	"errors"

	"github.com/raitonoberu/sptlrx/player"
)

func New() (*Client, error) {
	return &Client{}, nil
}

// Client implements player.Player
type Client struct{}

func (c *Client) State() (*player.State, error) {
	return nil, errors.New("spotify is not supported (yet!)")
}
