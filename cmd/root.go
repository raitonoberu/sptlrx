package cmd

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"sptlrx/cookie"
	"sptlrx/spotify"
	"sptlrx/ui"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	gloss "github.com/charmbracelet/lipgloss"
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

	FlagHAlignment string
)

var rootCmd = &coral.Command{
	Use:          "sptlrx",
	Short:        "Spotify lyrics in your terminal",
	Long:         "A CLI app that shows time-synced Spotify lyrics in your terminal.",
	Version:      "v0.2.0",
	SilenceUsage: true,

	RunE: func(cmd *coral.Command, args []string) error {
		var clientCookie string

		if FlagCookie != "" {
			clientCookie = FlagCookie
		} else if envCookie := os.Getenv("SPOTIFY_COOKIE"); envCookie != "" {
			clientCookie = envCookie
		} else {
			fileCookie, err := cookie.Load()
			if err != nil {
				return fmt.Errorf("couldn't load cookie: %w", err)
			}
			clientCookie = fileCookie
		}

		if clientCookie == "" {
			fmt.Print(banner)
			fmt.Printf("Cookie will be stored in %s\n", cookie.Directory)
			fmt.Print(help)
			ask("Enter your cookie:", &clientCookie)
			fmt.Println("You can always clear cookie by running 'sptlrx clear'.")
		}

		client, err := spotify.NewClient(clientCookie)
		if err != nil {
			return fmt.Errorf("couldn't create client: %w", err)
		}
		if err := cookie.Save(clientCookie); err != nil {
			return fmt.Errorf("couldn't save cookie: %w", err)
		}

		hAlignment := 0.5
		switch FlagHAlignment {
		case "left":
			hAlignment = 0
		case "right":
			hAlignment = 1
		}

		p := tea.NewProgram(
			&ui.Model{
				Client:       client,
				HAlignment:   gloss.Position(hAlignment),
				StyleBefore:  parseStyle(FlagStyleBefore),
				StyleCurrent: parseStyle(FlagStyleCurrent),
				StyleAfter:   parseStyle(FlagStyleAfter),
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

func parseStyle(value string) gloss.Style {
	var style gloss.Style

	if value == "" {
		return style
	}

	for _, part := range strings.Split(value, ",") {
		switch part {
		case "bold":
			style = style.Bold(true)
		case "italic":
			style = style.Italic(true)
		case "underline":
			style = style.Underline(true)
		case "strikethrough":
			style = style.Strikethrough(true)
		case "blink":
			style = style.Blink(true)
		case "faint":
			style = style.Faint(true)
		default:
			if validateColor(part) {
				if style.GetForeground() == (gloss.NoColor{}) {
					style = style.Foreground(gloss.Color(part))
				} else {
					style = style.Background(gloss.Color(part))
					style.ColorWhitespace(false)
				}
			} else {
				fmt.Println("Invalid style:", part)
			}
		}
	}
	return style
}

func validateColor(color string) bool {
	if _, err := strconv.Atoi(color); err == nil {
		return true
	}
	if strings.HasPrefix(color, "#") {
		return true
	}
	return false
}

func init() {
	rootCmd.Flags().StringVar(&FlagCookie, "cookie", "", "your cookie")

	rootCmd.Flags().StringVar(&FlagStyleBefore, "before", "bold", "style of the lines before the current ones")
	rootCmd.Flags().StringVar(&FlagStyleCurrent, "current", "bold", "style of the current lines")
	rootCmd.Flags().StringVar(&FlagStyleAfter, "after", "faint", "style of the lines after the current ones")

	rootCmd.Flags().StringVar(&FlagHAlignment, "halign", "center", "initial horizontal alignment (left/center/right)")

	rootCmd.AddCommand(clearCmd)
	rootCmd.AddCommand(pipeCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
