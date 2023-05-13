package config

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"sptlrx/player"
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
	Cookie         string `yaml:"cookie"`
	Player         string `default:"spotify" yaml:"player"`
	Host           string `default:"lyricsapi.vercel.app" yaml:"host"`
	IgnoreErrors   bool   `default:"true" yaml:"ignoreErrors"`
	TimerInterval  int    `default:"200" yaml:"timerInterval"`
	UpdateInterval int    `default:"2000" yaml:"updateInterval"`

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

	Web struct {
		Port      uint16 `default:"0" yaml:"port"`
		NoBrowser bool   `default:"false" yaml:"noBrowser"`
	} `yaml:"web"`

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
}

func New() *Config {
	var config = &Config{}
	defaults.Set(config)
	return config
}

func Load() (*Config, error) {
	file, err := os.Open(Path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// workaround for compatibility with old versions
			cookiePath := path.Join(Directory, "cookie.txt")
			if cookieFile, err := os.Open(cookiePath); err == nil {
				b, err := ioutil.ReadAll(cookieFile)
				cookieFile.Close()

				os.Remove(cookiePath)

				if err == nil && b != nil {
					config := New()
					config.Cookie = string(b)
					Save(config)

					return config, nil
				}
			}
		}
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
	Undeline      bool `yaml:"undeline"`
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
	if s.Undeline {
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

// GetPlayer returns a player based on config values
func GetPlayer(conf *Config) (player.Player, error) {
	switch conf.Player {
	case "spotify":
		if conf.Cookie == "" {
			return nil, spotify.ErrInvalidCookie
		}
		return spotify.New(conf.Cookie)
	case "mpd":
		return mpd.New(conf.Mpd.Address, conf.Mpd.Password)
	case "mopidy":
		return mopidy.New(conf.Mopidy.Address)
	case "mpris":
		return mpris.New(conf.Mpris.Players)
	}
	return nil, fmt.Errorf("unknown player: \"%s\"", conf.Player)
}
