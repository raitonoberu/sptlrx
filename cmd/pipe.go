package cmd

import (
	"errors"
	"fmt"
	"os"
	"sptlrx/config"
	"sptlrx/pool"
	"sptlrx/spotify"
	"strings"

	"github.com/muesli/coral"
	"github.com/muesli/reflow/wordwrap"
	"github.com/muesli/reflow/wrap"
)

var (
	FlagLength       int
	FlagOverflow     string
	FlagIgnoreErrors bool
)

var pipeCmd = &coral.Command{
	Use:   "pipe",
	Short: "Start printing the current lines to stdout",

	RunE: func(cmd *coral.Command, args []string) error {
		var conf *config.Config

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

		if conf.Cookie == "" {
			return errors.New("couldn't find cookie")
		}

		client, err := spotify.NewClient(conf.Cookie)
		if err != nil {
			return fmt.Errorf("couldn't create client: %w", err)
		}

		if cmd.Flags().Changed("length") {
			conf.Pipe.Length = FlagLength
		}
		if cmd.Flags().Changed("overflow") {
			conf.Pipe.Overflow = FlagOverflow
		}
		if cmd.Flags().Changed("ignore-errors") {
			conf.Pipe.IgnoreErrors = FlagIgnoreErrors
		}

		if cmd.Flags().Changed("tinterval") {
			conf.TimerInterval = FlagTimerInterval
		}
		if cmd.Flags().Changed("uinterval") {
			conf.UpdateInterval = FlagUpdateInterval
		}

		ch := make(chan pool.Update)
		go pool.Listen(client, conf, ch)

		for update := range ch {
			if update.Err != nil {
				if !conf.Pipe.IgnoreErrors {
					fmt.Println(err.Error())
				}
				continue
			}
			if update.Lines == nil || !update.Lines.Timesynced() {
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

func init() {
	pipeCmd.Flags().IntVar(&FlagLength, "length", 0, "max length of line")
	pipeCmd.Flags().StringVar(&FlagOverflow, "overflow", "word", "how to wrap an overflowed line (none/word/ellipsis)")
	pipeCmd.Flags().BoolVar(&FlagIgnoreErrors, "ignore-errors", true, "don't print errors")
}
