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

	tea "github.com/charmbracelet/bubbletea"
	"github.com/muesli/coral"
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
)

var rootCmd = &coral.Command{
	Use:          "sptlrx",
	Short:        "Time-synced lyrics in your terminal",
	Long:         "A CLI app that shows time-synced lyrics in your terminal",
	Version:      "v1.0.0-rc1",
	SilenceUsage: true,

	RunE: func(cmd *coral.Command, args []string) error {
		if cmd.Flags().Changed("config") {
			// custom config path
			config.Path = FlagConfig
		}

		conf, err := config.Load()
		if err != nil {
			if !cmd.Flags().Changed("config") && errors.Is(err, os.ErrNotExist) {
				conf = config.New()
				fmt.Print(banner + "\n")
				fmt.Printf("Config will be stored in %s\n", config.Directory)
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
			provider = hosted.New()
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
		return p.Start()
	},
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&FlagCookie, "cookie", "c", "", "your cookie")
	rootCmd.PersistentFlags().StringVarP(&FlagPlayer, "player", "p", "spotify", "what player to use")
	rootCmd.PersistentFlags().StringVar(&FlagConfig, "config", config.Path, "path to config file")

	rootCmd.AddCommand(pipeCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
