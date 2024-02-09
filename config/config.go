package config

import (
	"fmt"
	"os"
	"path"
	"sptlrx/player"
	"sptlrx/services/browser"
	"sptlrx/services/mopidy"
	"sptlrx/services/mpd"
	"sptlrx/services/mpris"
	"sptlrx/services/spotify"
	"strconv"
	"strings"

	gloss "github.com/charmbracelet/lipgloss"
	"github.com/creasty/defaults"
	"gopkg.in/yaml.v2"
)

var Directory string
var Path string

func init() {
	d, err := os.UserConfigDir()
	if err != nil {
		panic(err)
	}
	Directory = path.Join(d, "sptlrx")
	Path = path.Join(Directory, "config.yaml")
}

type Config struct {
	Cookie         string   `yaml:"cookie"`
	Players        []string `default:"[spotify]" yaml:"players"`
	Host           string   `default:"lyricsapi.vercel.app" yaml:"host"`
	IgnoreErrors   bool     `default:"true" yaml:"ignoreErrors"`
	TimerInterval  int      `default:"200" yaml:"timerInterval"`
	UpdateInterval int      `default:"2000" yaml:"updateInterval"`

	Style struct {
		HAlignment string `default:"center" yaml:"hAlignment"`

		Before  Style `default:"{\"bold\": true}" yaml:"before"`
		Current Style `default:"{\"bold\": true}" yaml:"current"`
		After   Style `default:"{\"faint\": true}" yaml:"after"`
	} `yaml:"style"`

	Pipe struct {
		Length   int    `yaml:"length"`
		Overflow string `default:"word" yaml:"overflow"`
	} `yaml:"pipe"`

	Mpd struct {
		Address  string `default:"127.0.0.1:6600" yaml:"address"`
		Password string `yaml:"password"`
	} `yaml:"mpd"`

	Mopidy struct {
		Address string `default:"127.0.0.1:6680" yaml:"address"`
	} `yaml:"mopidy"`

	Mpris struct {
		Players []string `default:"[]" yaml:"players"`
	} `yaml:"mpris"`

	Browser struct {
		Port int `default:"8974" yaml:"port"`
	} `yaml:"browser"`

	Local struct {
		Folder string `yaml:"folder"`
	} `yaml:"local"`
}

func New() *Config {
	var config = &Config{}
	defaults.Set(config)
	return config
}

func Load() (*Config, error) {
	file, err := os.Open(Path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var config = &Config{}
	err = yaml.NewDecoder(file).Decode(config)
	return config, err
}

func Save(config *Config) error {
	err := os.MkdirAll(Directory, os.ModePerm)
	if err != nil {
		return err
	}

	file, err := os.Create(Path)
	if err != nil {
		return err
	}
	defer file.Close()

	return yaml.NewEncoder(file).Encode(config)
}

// https://stackoverflow.com/a/56080478
func (c *Config) UnmarshalYAML(f func(interface{}) error) error {
	defaults.Set(c)

	type plain Config
	if err := f((*plain)(c)); err != nil {
		return err
	}

	return nil
}

type Style struct {
	Background string `yaml:"background"`
	Foreground string `yaml:"foreground"`

	Bold          bool `yaml:"bold"`
	Italic        bool `yaml:"italic"`
	Underline      bool `yaml:"underline"`
	Strikethrough bool `yaml:"strikethrough"`
	Blink         bool `yaml:"blink"`
	Faint         bool `yaml:"faint"`
}

func (s Style) Parse() gloss.Style {
	var style gloss.Style
	if s.Background != "" && validateColor(s.Background) {
		style = style.Background(gloss.Color(s.Background))
		style.ColorWhitespace(false)
	}
	if s.Foreground != "" && validateColor(s.Foreground) {
		style = style.Foreground(gloss.Color(s.Foreground))
	}

	if s.Bold {
		style = style.Bold(true)
	}
	if s.Italic {
		style = style.Italic(true)
	}
	if s.Underline {
		style = style.Underline(true)
	}
	if s.Strikethrough {
		style = style.Strikethrough(true)
	}
	if s.Blink {
		style = style.Blink(true)
	}
	if s.Faint {
		style = style.Faint(true)
	}

	return style
}

func validateColor(color string) bool {
	if _, err := strconv.Atoi(color); err == nil {
		return true
	}
	if strings.HasPrefix(color, "#") {
		return true
	}
	return false
}

// GetPlayers returns a player based on config values
func GetPlayers(conf *Config) ([]*player.Player, error) {
	var players []*player.Player
	for _, p := range conf.Players {
		os.Stderr.WriteString("LOADR: Processing " + p)
		switch p {
		case "spotify":
			spotifyPlayer, _ := spotify.NewPlayer(conf.Cookie)
			players = append(players, &spotifyPlayer)
		case "mpd":
			mpdPlayer := mpd.New(conf.Mpd.Address, conf.Mpd.Password)
			players = append(players, &mpdPlayer)
		case "mopidy":
			mopidyPlayer := mopidy.New(conf.Mopidy.Address)
			players = append(players, &mopidyPlayer)
		case "mpris":
			mprisPlayer, _ := mpris.New(conf.Mpris.Players)
			players = append(players, &mprisPlayer)
		case "browser":
			browserPlayer, _ := browser.New(conf.Browser.Port)
			players = append(players, &browserPlayer)
		}
	}
	if len(players) > 0 {
		return players, nil
	}

	return nil, fmt.Errorf("unknown players: \"%s\"", conf.Players)
}
