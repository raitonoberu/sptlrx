package config

import (
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/raitonoberu/sptlrx/player"
	"github.com/raitonoberu/sptlrx/services/browser"
	"github.com/raitonoberu/sptlrx/services/mopidy"
	"github.com/raitonoberu/sptlrx/services/mpd"
	"github.com/raitonoberu/sptlrx/services/mpris"
	"github.com/raitonoberu/sptlrx/services/spotify"

	gloss "github.com/charmbracelet/lipgloss"
	"github.com/creasty/defaults"
	"gopkg.in/yaml.v2"
)

var (
	Directory string
	Path      string
)

func init() {
	d, err := os.UserConfigDir()
	if err != nil {
		panic(err)
	}
	Directory = path.Join(d, "sptlrx")
	Path = path.Join(Directory, "config.yaml")
}

type Config struct {
	Player         string `default:"spotify" yaml:"player"`
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
	config := &Config{}
	defaults.Set(config)
	return config
}

func Load() (*Config, error) {
	file, err := os.Open(Path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	config := &Config{}
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
	Underline     bool `yaml:"underline"`
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

// GetPlayer returns a player based on config values.
// It supports a comma-separated list in conf.Player for backward-compatible multi-player configuration.
// Example: "mopidy ,mpris" will try Mopidy first, then MPRIS if Mopidy init fails.
// Only players explicitly listed by the user will be tried.
func GetPlayer(conf *Config) (player.Player, error) {
	// Split by comma to support multi-player configuration while keeping single value compatible
	raw := strings.Split(conf.Player, ",")
	// If no comma present, raw will contain the original single value

	var errs []string
	for _, name := range raw {
		n := strings.TrimSpace(name)
		if n == "" {
			continue
		}

		switch n {
		case "spotify":
			p, err := spotify.New()
			if err != nil {
				errs = append(errs, fmt.Sprintf("spotify: %v", err))
				continue
			}
			return p, nil
		case "mpd":
			p, err := mpd.New(conf.Mpd.Address, conf.Mpd.Password)
			if err != nil {
				errs = append(errs, fmt.Sprintf("mpd: %v", err))
				continue
			}
			return p, nil
		case "mopidy":
			p, err := mopidy.New(conf.Mopidy.Address)
			if err != nil {
				errs = append(errs, fmt.Sprintf("mopidy: %v", err))
				continue
			}
			return p, nil
		case "mpris":
			p, err := mpris.New(conf.Mpris.Players)
			if err != nil {
				errs = append(errs, fmt.Sprintf("mpris: %v", err))
				continue
			}
			return p, nil
		case "browser":
			p, err := browser.New(conf.Browser.Port)
			if err != nil {
				errs = append(errs, fmt.Sprintf("browser: %v", err))
				continue
			}
			return p, nil
		default:
			// Unknown player name; record and continue to allow next specified players
			errs = append(errs, fmt.Sprintf("unknown player: %q", n))
			continue
		}
	}

	// If nothing succeeded, return aggregated error for easier troubleshooting
	if len(errs) == 0 {
		return nil, fmt.Errorf("no player specified")
	}
	return nil, fmt.Errorf("no player initialized successfully; tried: %s", strings.Join(errs, "; "))
}
