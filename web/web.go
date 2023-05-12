package web

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"os/exec"
	"runtime"
	"sptlrx/config"
	"sptlrx/lyrics"
	"sptlrx/pool"
	"sync"

	"github.com/labstack/echo/v4"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

//go:embed frontend/dist
var frontendFS embed.FS

type message struct {
	Lines   []lyrics.Line `json:"lines,omitempty"`
	Index   *int          `json:"index,omitempty"`
	Playing *bool         `json:"playing,omitempty"`
	Err     string        `json:"err,omitempty"`
}

type Server struct {
	Config  *config.Config
	Channel chan pool.Update

	wsMutex sync.RWMutex
	wsPool  map[*websocket.Conn]struct{}

	lines   []lyrics.Line
	index   int
	playing bool
	err     error
}

func (s *Server) Start() error {
	e := echo.New()

	staticFS, _ := fs.Sub(frontendFS, "frontend/dist")
	staticHandler := http.FileServer(http.FS(staticFS))
	e.GET("/*", echo.WrapHandler(staticHandler))

	e.GET("/ws", s.wsHandler)

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", s.Config.Web.Port))
	if err != nil {
		return err
	}
	e.Listener = listener

	port := listener.Addr().(*net.TCPAddr).Port
	fmt.Println("HTTP server started on port", port)

	if !s.Config.Web.NoBrowser {
		openInBrowser(fmt.Sprintf("http://localhost:%d", port))
	}

	s.wsPool = make(map[*websocket.Conn]struct{})

	go s.updateLoop()

	e.HideBanner = true
	e.HidePort = true
	return e.Start(fmt.Sprintf(":%d", port))
}

func (s *Server) wsHandler(c echo.Context) error {
	conn, err := websocket.Accept(c.Response().Writer, c.Request(), nil)
	if err != nil {
		return err
	}

	s.wsMutex.Lock()
	s.wsPool[conn] = struct{}{}

	defer func(conn *websocket.Conn) {
		s.wsMutex.Lock()
		delete(s.wsPool, conn)
		s.wsMutex.Unlock()
	}(conn)

	s.wsMutex.Unlock()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s.sendState(conn)

	conn.Reader(ctx)
	conn.Close(websocket.StatusPolicyViolation, "unexpected data message")

	return nil
}

func (s *Server) sendState(conn *websocket.Conn) error {
	msg := message{
		Lines:   s.lines,
		Index:   &s.index,
		Playing: &s.playing,
	}
	if s.err != nil {
		msg.Err = s.err.Error()
	}
	return wsjson.Write(context.Background(), conn, msg)
}

func (s *Server) updateLoop() {
	for {
		update := <-s.Channel

		msg := message{}
		if lyricsChanged(s.lines, update.Lines) {
			s.lines = update.Lines
			msg.Lines = update.Lines

			if len(update.Lines) == 0 {
				// track is over, hiding lyrics
				index := -1
				msg.Index = &index
			}
		}

		if s.index != update.Index {
			s.index = update.Index
			msg.Index = &update.Index
		}

		if s.playing != update.Playing {
			s.playing = update.Playing
			msg.Playing = &update.Playing
		}

		// TODO: does this work?
		if s.err != update.Err {
			s.err = update.Err
			if update.Err != nil {
				msg.Err = update.Err.Error()
			}
		}

		s.notifyAll(msg)
	}
}

func (s *Server) notifyAll(m message) {
	s.wsMutex.RLock()
	defer s.wsMutex.RUnlock()

	for conn := range s.wsPool {
		// TODO: timeout?
		err := wsjson.Write(context.Background(), conn, m)
		if err != nil {
			fmt.Println(err)
		}
	}
}

func openInBrowser(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default:
		cmd = "xdg-open"
	}
	args = append(args, url)
	return exec.Command(cmd, args...).Start()
}

func lyricsChanged(s1, s2 []lyrics.Line) bool {
	if len(s1) != len(s2) {
		return true
	}
	return len(s1) != 0 && &s1[0] != &s2[0]
}
