package cmd

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"sptlrx/config"
	"sptlrx/spotify"
	"sptlrx/ui"
	"strings"

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

const help = `
How to get setup:

  1. Open your browser.
  2. Press F12, open the 'Network' tab and go to open.spotify.com.
  3. Click on the first request to open.spotify.com.
  4. Scroll down to the 'Request Headers', right click the 'cookie' field and select 'Copy value'.
`

var (
	FlagCookie string

	FlagStyleBefore  string
	FlagStyleCurrent string
	FlagStyleAfter   string
	FlagHAlignment   string

	FlagTimerInterval  int
	FlagUpdateInterval int
)

var rootCmd = &coral.Command{
	Use:          "sptlrx",
	Short:        "Spotify lyrics in your terminal",
	Long:         "A CLI app that shows time-synced Spotify lyrics in your terminal.",
	Version:      "v0.2.0",
	SilenceUsage: true,

	RunE: func(cmd *coral.Command, args []string) error {
		var conf *config.Config

		conf, err := config.Load()
		if err != nil {
			return fmt.Errorf("couldn't load config: %w", err)
		}
		if conf == nil {
			conf = config.New()
			fmt.Print(banner)
			fmt.Printf("Config will be stored in %s\n", config.Directory)
			config.Save(conf)
		}

		if FlagCookie != "" {
			conf.Cookie = FlagCookie
		} else if envCookie := os.Getenv("SPOTIFY_COOKIE"); envCookie != "" {
			conf.Cookie = envCookie
		}

		if conf.Cookie == "" {
			fmt.Print(help)
			ask("Enter your cookie:", &conf.Cookie)
			config.Save(conf)
		}

		client, err := spotify.NewClient(conf.Cookie)
		if err != nil {
			return fmt.Errorf("couldn't create client: %w", err)
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

		if cmd.Flags().Changed("tinterval") {
			conf.TimerInterval = FlagTimerInterval
		}
		if cmd.Flags().Changed("uinterval") {
			conf.UpdateInterval = FlagUpdateInterval
		}

		p := tea.NewProgram(
			&ui.Model{
				Client: client,
				Config: conf,
			},
			tea.WithAltScreen(),
		)
		return p.Start()
	},
}

func ask(what string, answer *string) {
	var ok bool
	scanner := bufio.NewScanner(os.Stdin)
	for !ok {
		fmt.Println("\n" + what)
		scanner.Scan()

		if err := scanner.Err(); err != nil {
			log.Fatal(err)
		}

		line := strings.TrimSpace(scanner.Text())

		if line != "" {
			ok = true
			*answer = line
		} else {
			fmt.Println("The value can't be empty.")
		}
	}
}

func parseStyleFlag(value string) config.StyleConfig {
	var style config.StyleConfig

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
	rootCmd.PersistentFlags().StringVar(&FlagCookie, "cookie", "", "your cookie")

	rootCmd.Flags().StringVar(&FlagStyleBefore, "before", "bold", "style of the lines before the current ones")
	rootCmd.Flags().StringVar(&FlagStyleCurrent, "current", "bold", "style of the current lines")
	rootCmd.Flags().StringVar(&FlagStyleAfter, "after", "faint", "style of the lines after the current ones")
	rootCmd.Flags().StringVar(&FlagHAlignment, "halign", "center", "initial horizontal alignment (left/center/right)")

	rootCmd.PersistentFlags().IntVar(&FlagTimerInterval, "tinterval", 200, "interval for the internal timer (ms)")
	rootCmd.PersistentFlags().IntVar(&FlagUpdateInterval, "uinterval", 200, "interval for updating playback status (ms)")

	rootCmd.AddCommand(pipeCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
