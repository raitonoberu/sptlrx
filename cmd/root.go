package cmd

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"sptlrx/cookie"
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

var rootCmd = &coral.Command{
	Use:          "sptlrx",
	Short:        "Spotify lyrics in your terminal",
	Long:         "A CLI app that shows time-synced Spotify lyrics in your terminal.",
	Version:      "v0.1.0",
	SilenceUsage: true,

	RunE: func(cmd *coral.Command, args []string) error {
		var err error
		// check env first
		c := os.Getenv("SPOTIFY_COOKIE")
		if c == "" {
			// try loading cookie from file
			c, err = cookie.Load()
			if err != nil {
				return fmt.Errorf("couldn't load cookie: %w", err)
			}
		}

		if c == "" {
			fmt.Print(banner)
			fmt.Printf("Cookie will be stored in %s\n", cookie.Directory)
			fmt.Print(help)
			ask("Enter your cookie:", &c)
			fmt.Println("You can always clear cookie by running 'sptlrx clear'.")
		}

		client, err := spotify.NewClient(c)
		if err != nil {
			return fmt.Errorf("couldn't create client: %w", err)
		}
		if err := cookie.Save(c); err != nil {
			return fmt.Errorf("couldn't save cookie: %w", err)
		}

		p := tea.NewProgram(
			ui.NewModel(client),
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

func init() {
	rootCmd.AddCommand(clearCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
