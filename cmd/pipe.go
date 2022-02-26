package cmd

import (
	"errors"
	"fmt"
	"os"
	"sptlrx/cookie"
	"sptlrx/pool"
	"sptlrx/spotify"
	"strings"

	"github.com/muesli/coral"
	"github.com/muesli/reflow/wordwrap"
	"github.com/muesli/reflow/wrap"
)

var (
	FlagLength   int
	FlagOverflow string
)

var pipeCmd = &coral.Command{
	Use:   "pipe",
	Short: "Pipes the current line to stdout",

	RunE: func(cmd *coral.Command, args []string) error {
		var clientCookie string

		if FlagCookie != "" {
			clientCookie = FlagCookie
		} else if envCookie := os.Getenv("SPOTIFY_COOKIE"); envCookie != "" {
			clientCookie = envCookie
		} else {
			clientCookie, _ = cookie.Load()
		}

		if clientCookie == "" {
			return errors.New("couldn't find cookie")
		}

		client, err := spotify.NewClient(clientCookie)
		if err != nil {
			return fmt.Errorf("couldn't create client: %w", err)
		}
		if err := cookie.Save(clientCookie); err != nil {
			return fmt.Errorf("couldn't save cookie: %w", err)
		}

		ch := make(chan pool.Update)
		go pool.Listen(client, ch)

		for update := range ch {
			if update.Err != nil {
				fmt.Println(err.Error())
			}
			if update.Lines != nil && update.Lines.Timesynced() {
				line := update.Lines[update.Index].Words
				if FlagLength == 0 {
					fmt.Println(line)
				} else {
					// TODO: find out if there is a better way to cut the line
					switch FlagOverflow {
					case "word":
						s := wordwrap.String(line, FlagLength)
						fmt.Println(strings.Split(s, "\n")[0])
					case "none":
						s := wrap.String(line, FlagLength)
						fmt.Println(strings.Split(s, "\n")[0])
					case "ellipsis":
						s := wrap.String(line, FlagLength)
						lines := strings.Split(s, "\n")
						if len(lines) == 1 {
							fmt.Println(lines[0])
						} else {
							s := wrap.String(lines[0], FlagLength-3)
							fmt.Println(strings.Split(s, "\n")[0] + "...")
						}
					}
				}
			} else {
				// no lyrics or not timesynced
				fmt.Println("")
			}
		}
		return nil
	},
}

func init() {
	pipeCmd.Flags().StringVar(&FlagCookie, "cookie", "", "your cookie")
	pipeCmd.Flags().IntVar(&FlagLength, "length", 0, "max length of line")
	pipeCmd.Flags().StringVar(&FlagOverflow, "overflow", "word", "how to wrap an overflowed line (none/word/ellipsis)")
}
