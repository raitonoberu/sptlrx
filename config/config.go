package config

import (
	"errors"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"strings"

	gloss "github.com/charmbracelet/lipgloss"
	"github.com/creasty/defaults"
	"gopkg.in/yaml.v2"
)

var Directory string

func init() {
	d, err := os.UserConfigDir()
	if err != nil {
		panic(err)
	}
	Directory = path.Join(d, "sptlrx")
}

type Config struct {
	Cookie string `yaml:"cookie"`
	// Player         string `default:"spotify" yaml:"player"`
	TimerInterval  int `default:"200" yaml:"timerInterval"`
	UpdateInterval int `default:"3000" yaml:"updateInterval"`

	Style struct {
		HAlignment string `default:"center" yaml:"hAlignment"`

		Before  StyleConfig `default:"{\"bold\": true}" yaml:"before"`
		Current StyleConfig `default:"{\"bold\": true}" yaml:"current"`
		After   StyleConfig `default:"{\"faint\": true}" yaml:"after"`
	} `yaml:"style"`

	Pipe struct {
		Length       int    `yaml:"length"`
		Overflow     string `default:"word" yaml:"overflow"`
		IgnoreErrors bool   `default:"true" yaml:"ignoreErrors"`
	} `yaml:"pipe"`

	// Mpd struct {
	// 	Hostname string `default:"127.0.0.1" yaml:"hostname"`
	// 	Port     int    `default:"6600" yaml:"port"`
	// 	Password string `yaml:"password"`
	// } `yaml:"mpd"`

	// Mopidy struct {
	// 	Hostname string `default:"127.0.0.1" yaml:"hostname"`
	// 	Port     int    `default:"6680" yaml:"port"`
	// } `yaml:"mopidy"`
}

func New() *Config {
	var config = &Config{}
	defaults.Set(config)
	return config
}

func Load() (*Config, error) {
	file, err := os.Open(path.Join(Directory, "config.yaml"))
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return nil, err
		}

		// workaround for compatibility with old versions
		if cookieFile, err := os.Open(path.Join(Directory, "cookie.txt")); err == nil {
			b, err := ioutil.ReadAll(cookieFile)
			cookieFile.Close()

			os.Remove(path.Join(Directory, "cookie.txt"))

			if err == nil && b != nil {
				config := New()
				config.Cookie = string(b)
				Save(config)

				return config, nil
			}
		}
		return nil, nil
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

	file, err := os.Create(path.Join(Directory, "config.yaml"))
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

type StyleConfig struct {
	Background string `yaml:"background"`
	Foreground string `yaml:"foreground"`

	Bold          bool `yaml:"bold"`
	Italic        bool `yaml:"italic"`
	Undeline      bool `yaml:"undeline"`
	Strikethrough bool `yaml:"strikethrough"`
	Blink         bool `yaml:"blink"`
	Faint         bool `yaml:"faint"`
}

func (s StyleConfig) Parse() gloss.Style {
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
