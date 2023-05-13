package cmd

import (
	"fmt"
	"os"
	"sptlrx/config"
	"sptlrx/lyrics"
	"sptlrx/pool"
	"sptlrx/services/hosted"
	"sptlrx/services/spotify"
	"strings"

	"github.com/muesli/reflow/wordwrap"
	"github.com/muesli/reflow/wrap"
	"github.com/spf13/cobra"
)

var pipeCmd = &cobra.Command{
	Use:   "pipe",
	Short: "Start printing the current lines to stdout",

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

		var ch = make(chan pool.Update)
		go pool.Listen(player, provider, conf, ch)

		for update := range ch {
			if update.Err != nil {
				if !conf.Pipe.IgnoreErrors {
					fmt.Println(update.Err.Error())
				}
				continue
			}
			if update.Lines == nil || !lyrics.Timesynced(update.Lines) {
				fmt.Println("")
				continue
			}
			line := update.Lines[update.Index].Words
			if conf.Pipe.Length == 0 {
				fmt.Println(line)
			} else {
				switch conf.Pipe.Overflow {
				case "word":
					s := wordwrap.String(line, conf.Pipe.Length)
					fmt.Println(strings.Split(s, "\n")[0])
				case "none":
					s := wrap.String(line, conf.Pipe.Length)
					fmt.Println(strings.Split(s, "\n")[0])
				case "ellipsis":
					s := wrap.String(line, conf.Pipe.Length)
					lines := strings.Split(s, "\n")
					if len(lines) == 1 {
						fmt.Println(lines[0])
					} else {
						s := wrap.String(lines[0], conf.Pipe.Length-3)
						fmt.Println(strings.Split(s, "\n")[0] + "...")
					}
				}
			}
		}
		return nil
	},
}
