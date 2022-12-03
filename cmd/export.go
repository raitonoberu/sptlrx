package cmd

import (
	"fmt"
	"os"
	"sptlrx/config"
	"sptlrx/lyrics"
	"sptlrx/pool"
	"sptlrx/services/hosted"
	"sptlrx/services/spotify"
	"time"

	"github.com/muesli/coral"
)

var exportCmd = &coral.Command{
	Use:   "export",
	Short: "Export lyrics to stdout in LRC format.",

	RunE: func(cmd *coral.Command, args []string) error {
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
			provider = hosted.New()
		}

		var ch = make(chan pool.Update)
		go pool.Listen(player, provider, conf, ch)

		update, ok := <-ch
		if err := update.Err; !ok || err != nil || update.Lines == nil || !lyrics.Timesynced(update.Lines) {
			return err
		}
		for _, line := range update.Lines {
			fmt.Print(toLRCTime(line.Time), line.Words, "\n")
		}
		return nil
	},
}

func toLRCTime(t int) string {
	d := time.Duration(t-1).Abs() * time.Millisecond
	const width2 = "%2.2d"
	mm, d := fmt.Sprintf(width2, d/time.Minute), d%time.Minute
	ss, d := fmt.Sprintf(width2, d/time.Second), d%time.Second
	xx, _ := fmt.Sprintf(width2, d/(time.Second/100)), d%(time.Second/100)
	if len(mm) > 2 || len(ss) > 2 || len(xx) > 2 {
		mm, ss, xx = "99", "59", "99"
	}
	return "[" + mm + ":" + ss + "." + xx + "]"
}
