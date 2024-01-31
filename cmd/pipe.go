package cmd

import (
	"fmt"
	"sptlrx/config"
	"sptlrx/lyrics"
	"sptlrx/pool"
	"strings"

	"github.com/muesli/reflow/wordwrap"
	"github.com/muesli/reflow/wrap"
	"github.com/spf13/cobra"
)

var pipeCmd = &cobra.Command{
	Use:   "pipe",
	Short: "Start printing the current lines to stdout",

	RunE: func(cmd *cobra.Command, args []string) error {
		conf, err := loadConfig(cmd)
		if err != nil {
			return fmt.Errorf("couldn't load config: %w", err)
		}
		players, err := loadPlayers(conf)
		if err != nil {
			return fmt.Errorf("couldn't load players: %w", err)
		}
		provider, err := loadProvider(conf)
		if err != nil {
			return fmt.Errorf("couldn't load provider: %w", err)
		}

		ch := make(chan pool.Update)
		go pool.Listen(players, provider, conf, ch)

		for update := range ch {
			printUpdate(update, conf)
		}
		return nil
	},
}

func printUpdate(update pool.Update, conf *config.Config) {
	if update.Err != nil {
		if !conf.IgnoreErrors {
			fmt.Println(update.Err.Error())
		}
		return
	}
	if update.Lines == nil || !lyrics.Timesynced(update.Lines) {
		fmt.Println("")
		return
	}
	line := update.Lines[update.Index].Words
	if conf.Pipe.Length == 0 {
		fmt.Println(line)
		return
	}
	switch conf.Pipe.Overflow {
	case "none":
		s := wrap.String(line, conf.Pipe.Length)
		fmt.Println(strings.Split(s, "\n")[0])
	case "word":
		s := wordwrap.String(line, conf.Pipe.Length)
		fmt.Println(strings.Split(s, "\n")[0])
	case "ellipsis":
		s := wrap.String(line, conf.Pipe.Length)
		lines := strings.Split(s, "\n")
		if len(lines) == 1 {
			fmt.Println(lines[0])
			return
		}
		s = wrap.String(lines[0], conf.Pipe.Length-3)
		fmt.Println(strings.Split(s, "\n")[0] + "...")
	}
}
