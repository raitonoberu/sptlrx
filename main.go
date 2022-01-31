package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"sptlrx/cookie"
	"sptlrx/spotify"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
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

func main() {
	if len(os.Args) > 1 {
		if os.Args[1] == "-c" || os.Args[1] == "--clear" {
			cookie.Clear()
		} else {
			fmt.Println("Unknown command:", os.Args[1])
			return
		}
	}

	c := os.Getenv("SPOTIFY_COOKIE")
	if c != "" {
		fmt.Println("Using cookie from the $SPOTIFY_COOKIE enviroment variable.")
	} else {
		// try loading cookie from file
		c = cookie.Load()
	}

	if c == "" {
		fmt.Print(banner)
		fmt.Printf("Cookie will be stored in %s\n", cookie.Directory)
		fmt.Print(help)
		ask("Enter your cookie:", &c)
		fmt.Println("You can always clear cookie by running 'sptlrx --clear'.")
	}

	sp, err := spotify.NewClient(c)
	if err != nil {
		fmt.Printf("Couldn't create client: %v\n", err)
		return
	}
	cookie.Save(c)
	client = sp

	p := tea.NewProgram(
		&model{
			hAlignment: 0.5,
		},
		tea.WithAltScreen(),
	)
	if err := p.Start(); err != nil {
		log.Fatal(err)
	}
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
