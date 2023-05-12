package cmd

import (
	"fmt"
	"os"
	"sptlrx/config"
	"sptlrx/lyrics"
	"sptlrx/pool"
	"sptlrx/services/hosted"
	"sptlrx/services/spotify"
	"sptlrx/web"

	"github.com/spf13/cobra"
)

var (
	FlagPort      uint16
	FlagNoBrowser bool
)

var webCmd = &cobra.Command{
	Use:   "web",
	Short: "Start web server to display lyrics in your browser",

	RunE: func(cmd *cobra.Command, args []string) error {
		if cmd.Flags().Changed("config") {
			// custom config path
			config.Path = FlagConfig
		}

		conf, err := config.Load()
		if err != nil {
			return fmt.Errorf("couldn't load config: %w", err)
		}

		if conf == nil {
			conf = config.New()
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

		if cmd.Flags().Changed("port") {
			conf.Web.Port = FlagPort
		}
		if cmd.Flags().Changed("no-browser") {
			conf.Web.NoBrowser = FlagNoBrowser
		}

		var ch = make(chan pool.Update)
		go pool.Listen(player, provider, conf, ch)
		server := &web.Server{
			Config:  conf,
			Channel: ch,
		}
		return server.Start()
	},
}

func init() {
	webCmd.Flags().Uint16Var(&FlagPort, "port", 0, "port to host the web server on (default: random)")
	webCmd.Flags().BoolVar(&FlagNoBrowser, "no-browser", false, "don't open the browser automatically")
}
