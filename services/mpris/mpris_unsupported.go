//go:build windows || darwin

package mpris

import (
	"errors"
	"github.com/raitonoberu/sptlrx/player"
)

func New(players []string) (*Client, error) {
	return nil, errors.New("darwin is not supported")
}

// Client implements player.Player
type Client struct{}

func (p *Client) State() (*player.State, error) {
	return nil, nil
}
