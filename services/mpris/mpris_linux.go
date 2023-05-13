package mpris

import (
	"sptlrx/player"
	"strings"

	"github.com/Pauloo27/go-mpris"
	"github.com/godbus/dbus/v5"
)

func New(players []string) (*Client, error) {
	conn, err := dbus.SessionBus()
	if err != nil {
		return nil, err
	}
	return &Client{players, conn}, nil
}

// Client implements player.Player
type Client struct {
	players []string
	conn    *dbus.Conn
}

func (c *Client) getPlayer() (*mpris.Player, error) {
	players, err := mpris.List(c.conn)
	if err != nil {
		return nil, err
	}
	if len(players) == 0 {
		return nil, nil
	}

	if len(c.players) == 0 {
		return mpris.New(c.conn, players[0]), nil
	}

	// iterating over configured names
	for _, p := range c.players {
		for _, player := range players {
			// trim "org.mpris.MediaPlayer2."
			if player[23:] == p {
				return mpris.New(c.conn, player), nil
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
		return nil, err
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
