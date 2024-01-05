package mpris

import (
	"path"
	"sptlrx/player"
	"strings"

	"github.com/Pauloo27/go-mpris"
	"github.com/godbus/dbus/v5"
)

func New(players []string) (*Client, error) {
	return &Client{players}, nil
}

// Client implements player.Player
type Client struct {
	players []string
}

func (c *Client) getPlayer() (*mpris.Player, error) {
	conn, err := dbus.SessionBus()
	if err != nil {
		return nil, err
	}

	players, err := mpris.List(conn)
	if err != nil {
		return nil, err
	}
	if len(players) == 0 {
		return nil, nil
	}

	if len(c.players) == 0 {
		return mpris.New(conn, players[0]), nil
	}

	// iterating over configured names
	for _, p := range c.players {
		for _, player := range players {
			// support pattern matching
			match, err := path.Match("org.mpris.MediaPlayer2."+p, player)
			if err != nil {
				return nil, err
			}
			if match {
				return mpris.New(conn, player), nil
			}
		}
	}
	return nil, nil
}

func (c *Client) State() (*player.State, error) {
	p, err := c.getPlayer()
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, nil
	}

	status, err := p.GetPlaybackStatus()
	if err != nil {
		return nil, err
	}
	position, err := p.GetPosition()
	if err != nil {
		// unsupported player
		return nil, nil
	}
	meta, err := p.GetMetadata()
	if err != nil {
		return nil, err
	}

	var title string
	if t, ok := meta["xesam:title"].Value().(string); ok {
		title = t
	}

	var artist string
	switch a := meta["xesam:artist"].Value(); a.(type) {
	case string:
		artist = a.(string)
	case []string:
		artist = strings.Join(a.([]string), " ")
	}

	var query string
	if artist != "" {
		query = artist + " " + title
	} else {
		query = title
	}

	return &player.State{
		ID:       query, // use query as id since mpris:trackid is broken
		Query:    query,
		Position: int(position * 1000), // secs to ms
		Playing:  status == mpris.PlaybackPlaying,
	}, err
}
