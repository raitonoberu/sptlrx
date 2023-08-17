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
	"time"

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
	Error   string        `json:"error,omitempty"`
}

type Server struct {
	Config  *config.Config
	Channel chan pool.Update

	wsMutex sync.RWMutex
	wsPool  map[*websocket.Conn]*sync.Mutex

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

	s.wsPool = make(map[*websocket.Conn]*sync.Mutex)

	go s.updateLoop()

	e.HideBanner = true
	e.HidePort = true
	return e.Start(fmt.Sprintf(":%d", port))
}

func (s *Server) wsHandler(c echo.Context) error {
	conn, err := websocket.Accept(c.Response(), c.Request(),
		&websocket.AcceptOptions{
			InsecureSkipVerify: true,
		})
	if err != nil {
		return err
	}

	s.wsMutex.Lock()
	s.wsPool[conn] = &sync.Mutex{}

	defer func(conn *websocket.Conn) {
		s.wsMutex.Lock()
		delete(s.wsPool, conn)
		s.wsMutex.Unlock()
	}(conn)

	s.wsMutex.Unlock()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s.sendInitialState(conn)

	conn.Reader(ctx)
	conn.Close(websocket.StatusPolicyViolation, "unexpected data message")

	return nil
}

func (s *Server) sendInitialState(conn *websocket.Conn) error {
	msg := message{
		Lines:   s.lines,
		Index:   &s.index,
		Playing: &s.playing,
	}
	if s.err != nil {
		msg.Error = s.err.Error()
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
				msg.Error = update.Err.Error()
			}
		}

		go s.notifyAll(msg)
	}
}

func (s *Server) notifyAll(m message) {
	s.wsMutex.RLock()
	wg := sync.WaitGroup{}

	for conn, mu := range s.wsPool {
		wg.Add(1)
		go func(conn *websocket.Conn, mu *sync.Mutex) {
			mu.Lock()

			ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
			err := wsjson.Write(ctx, conn, m)
			if err != nil && !s.Config.IgnoreErrors {
				fmt.Println(err)
			}
			cancel()

			mu.Unlock()
			wg.Done()
		}(conn, mu)
	}

	wg.Wait()
	s.wsMutex.RUnlock()
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
