package mpris

import (
	"errors"
	"sptlrx/player"
)

func New(players []string) (*Client, error) {
	return nil, errors.New("windows is not supported")
}

// Client implements player.Player
type Client struct{}

func (p *Client) State() (*player.State, error) {
	return nil, nil
}
