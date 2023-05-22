package cmd

import (
	"errors"
	"fmt"
	"os"
	"sptlrx/config"
	"sptlrx/lyrics"
	"sptlrx/pool"
	"sptlrx/services/hosted"
	"sptlrx/services/spotify"
	"sptlrx/ui"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

const banner = `
             _    _             
 ___  _ __  | |_ | | _ __ __  __
/ __|| '_ \ | __|| || '__|\ \/ /
\__ \| |_) || |_ | || |    >  < 
|___/| .__/  \__||_||_|   /_/\_\
     |_|                        
`

const help = `  1. Open your browser.
  2. Press F12, open the 'Network' tab and go to open.spotify.com.
  3. Click on the first request to open.spotify.com.
  4. Scroll down to the 'Request Headers', right click the 'cookie' field and select 'Copy value'.
  5. Paste it into your config file.`

var (
	FlagCookie string
	FlagPlayer string
	FlagConfig string

	FlagStyleBefore  string
	FlagStyleCurrent string
	FlagStyleAfter   string
	FlagHAlignment   string

	FlagVerbose bool
)

var rootCmd = &cobra.Command{
	Use:          "sptlrx",
	Short:        "Synchronized lyrics in your terminal",
	Long:         "A CLI app that shows time-synchronized lyrics in your terminal",
	Version:      "v1.0.0",
	SilenceUsage: true,

	RunE: func(cmd *cobra.Command, args []string) error {
		if cmd.Flags().Changed("config") {
			// custom config path
			config.Path = FlagConfig
		}

		conf, err := config.Load()
		if err != nil {
			if !cmd.Flags().Changed("config") && errors.Is(err, os.ErrNotExist) {
				conf = config.New()
				fmt.Print(banner + "\n")
				fmt.Printf("Config file location: %s\n", config.Path)
				config.Save(conf)
			} else {
				return fmt.Errorf("couldn't load config: %w", err)
			}
		}

		if FlagCookie != "" {
			conf.Cookie = FlagCookie
		} else if envCookie := os.Getenv("SPOTIFY_COOKIE"); envCookie != "" {
			conf.Cookie = envCookie
		}

		if cmd.Flags().Changed("player") {
			conf.Player = FlagPlayer
		}

		player, err := config.GetPlayer(conf)
		if err != nil {
			if errors.Is(err, spotify.ErrInvalidCookie) {
				fmt.Println("If you want to use Spotify as your player, you need to set up your cookie.")
				fmt.Println(help)
			}
			return err
		}

		var provider lyrics.Provider
		if conf.Cookie != "" {
			if spt, ok := player.(*spotify.Client); ok {
				// use existing spotify client
				provider = spt
			} else {
				// create new client
				client, err := spotify.New(conf.Cookie)
				if err != nil {
					return err
				}
				provider = client
			}
		} else {
			// use hosted provider
			provider = hosted.New(conf.Host)
		}

		if cmd.Flags().Changed("before") {
			conf.Style.Before = parseStyleFlag(FlagStyleBefore)
		}
		if cmd.Flags().Changed("current") {
			conf.Style.Current = parseStyleFlag(FlagStyleCurrent)
		}
		if cmd.Flags().Changed("after") {
			conf.Style.After = parseStyleFlag(FlagStyleAfter)
		}
		if cmd.Flags().Changed("halign") {
			conf.Style.HAlignment = FlagHAlignment
		}

		if FlagVerbose {
			conf.IgnoreErrors = false
		}

		var ch = make(chan pool.Update)
		go pool.Listen(player, provider, conf, ch)

		p := tea.NewProgram(
			&ui.Model{
				Channel: ch,
				Config:  conf,
			},
			tea.WithAltScreen(),
		)
		_, err = p.Run()
		return err
	},
}

func parseStyleFlag(value string) config.Style {
	var style config.Style

	for _, part := range strings.Split(value, ",") {
		switch part {
		case "bold":
			style.Bold = true
		case "italic":
			style.Italic = true
		case "underline":
			style.Undeline = true
		case "strikethrough":
			style.Strikethrough = true
		case "blink":
			style.Blink = true
		case "faint":
			style.Faint = true
		default:
			if style.Foreground == "" {
				style.Foreground = part
			} else if style.Background == "" {
				style.Background = part
			}
		}
	}
	return style
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&FlagCookie, "cookie", "c", "", "your cookie")
	rootCmd.PersistentFlags().StringVarP(&FlagPlayer, "player", "p", "spotify", "what player to use")
	rootCmd.PersistentFlags().StringVar(&FlagConfig, "config", config.Path, "path to config file")

	rootCmd.Flags().StringVar(&FlagStyleBefore, "before", "bold", "style of the lines before the current one")
	rootCmd.Flags().StringVar(&FlagStyleCurrent, "current", "bold", "style of the current line")
	rootCmd.Flags().StringVar(&FlagStyleAfter, "after", "faint", "style of the lines after the current one")
	rootCmd.Flags().StringVar(&FlagHAlignment, "halign", "center", "initial horizontal alignment (left/center/right)")

	rootCmd.PersistentFlags().BoolVarP(&FlagVerbose, "verbose", "v", false, "force print errors")

	rootCmd.AddCommand(pipeCmd)
	rootCmd.AddCommand(webCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
