//go:build !(windows || darwin)

package mpris

import (
	"sptlrx/player"
	"strings"
	"net/url"
	"path/filepath"

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

	// iterating over configured whitelisted players
	for _, p := range c.players {
		// adding the D-Bus bus name prefix
		p := "org.mpris.MediaPlayer2." + p
		for _, player := range players {
			// check for the name with and without the instance suffix
			if p == player || strings.HasPrefix(player, p+".instance") {
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

	var uri *url.URL
	// In case the player uses the file name with extension as title
	if u, ok := meta["xesam:url"].Value().(string); ok {
		u, err := url.Parse(u)
		if err == nil && u.Path != "" {
			ext := filepath.Ext(u.Path)
			uri = u
			// some players use filename as title when tag is absent => trim extension from title
			title = strings.TrimSuffix(title, ext)
		}
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
		Track: player.TrackMetadata{
			ID:    query, // use query as id since mpris:trackid is broken
			Uri:   uri,
			Query: query,
		},
		Position: int(position * 1000), // secs to ms
		Playing:  status == mpris.PlaybackPlaying,
	}, err
}
