package mpris

import (
	"sptlrx/player"
	"strings"

	"github.com/Pauloo27/go-mpris"
	"github.com/godbus/dbus/v5"
)

func New(name string) (*Client, error) {
	conn, err := dbus.SessionBus()
	if err != nil {
		return nil, err
	}
	return &Client{name, conn}, nil
}

// Client implements player.Player
type Client struct {
	name string
	conn *dbus.Conn
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func (p *Client) getPlayer() (*mpris.Player, error) {
	names, err := mpris.List(p.conn)
	if err != nil {
		return nil, err
	}
	if len(names) == 0 {
		return nil, nil
	}

	if len(p.name) == 0 {
		return mpris.New(p.conn, names[0]), nil
	}

	if !stringInSlice(p.name, names) {
		return nil, nil
	}
	return mpris.New(p.conn, p.name), nil
}

func (p *Client) State() (*player.State, error) {
	pl, err := p.getPlayer()
	if err != nil {
		return nil, err
	}
	if pl == nil {
		return nil, nil
	}

	status, err := pl.GetPlaybackStatus()
	if err != nil {
		return nil, err
	}
	position, err := pl.GetPosition()
	if err != nil {
		// unsupported player
		return nil, nil
	}
	meta, err := pl.GetMetadata()
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
