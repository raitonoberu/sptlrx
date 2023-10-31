package browser

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"sptlrx/player"
	"strconv"
	"strings"
	"sync"
	"time"

	"nhooyr.io/websocket"
)

const helloMessage = "ADAPTER_VERSION 1.0.0;WNPRLIB_REVISION 2"

type state int

const (
	stopped state = iota
	paused
	playing
)

func New(port int) (*Client, error) {
	c := &Client{}
	return c, c.start(port)
}

// Client implements player.Player
type Client struct {
	state    state
	position int
	title    string
	artist   string

	updateTime time.Time

	stateMu sync.Mutex
	connMu  sync.Mutex
}

func (c *Client) handler(w http.ResponseWriter, r *http.Request) {
	// make sure we only have one connection
	c.connMu.Lock()
	defer c.connMu.Unlock()

	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		InsecureSkipVerify: true,
	})
	if err != nil {
		return
	}
	defer conn.Close(websocket.StatusInternalError, "internal error")

	writer, err := conn.Writer(r.Context(), websocket.MessageText)
	if err != nil {
		return
	}

	writer.Write([]byte(helloMessage))
	writer.Close()

	for {
		t, reader, err := conn.Reader(r.Context())
		if err != nil {
			return
		}

		msg, err := io.ReadAll(reader)
		if err != nil {
			return
		}
		if t != websocket.MessageText || len(msg) == 0 {
			continue
		}
		c.processMessage(string(msg))
	}
}

func (c *Client) processMessage(msg string) {
	spaceIndex := strings.IndexByte(msg, ' ')
	if spaceIndex == -1 {
		return
	}

	msgType := strings.ToUpper(msg[:spaceIndex])
	data := msg[spaceIndex+1:]

	// we are not doing global locking here because
	// we are not interested in most of the messages
	switch msgType {
	case "STATE":
		c.stateMu.Lock()
		switch data {
		case "PLAYING":
			c.state = playing
		case "PAUSED":
			c.state = paused
		case "STOPPED":
			c.state = stopped
		}
		c.stateMu.Unlock()
	case "TITLE":
		c.stateMu.Lock()
		c.title = data
		c.stateMu.Unlock()
	case "ARTIST":
		c.stateMu.Lock()
		c.artist = data
		c.stateMu.Unlock()
	case "POSITION_SECONDS":
		pos, _ := strconv.Atoi(data)
		c.stateMu.Lock()
		c.position = pos * 1000
		c.updateTime = time.Now()
		c.stateMu.Unlock()
	}
}

func (c *Client) start(port int) error {
	l, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		return err
	}

	server := &http.Server{
		Handler: http.HandlerFunc(c.handler),
	}
	go server.Serve(l)
	return nil
}

func (c *Client) State() (*player.State, error) {
	c.stateMu.Lock()
	defer c.stateMu.Unlock()

	if c.state == stopped {
		return nil, nil
	}

	var query string
	if c.artist != "" {
		query = c.artist + " " + c.title
	} else {
		query = c.title
	}

	position := c.position
	if c.state != paused {
		position += int(time.Since(c.updateTime).Milliseconds())
	}
	return &player.State{
		ID:       query,
		Query:    query,
		Position: position,
		Playing:  c.state == playing,
	}, nil
}
