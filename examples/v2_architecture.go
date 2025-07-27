package main

import (
	"fmt"
	"log"

	"github.com/raitonoberu/sptlrx/services/lrclib"
	"github.com/raitonoberu/sptlrx/services/sources"
	"github.com/raitonoberu/sptlrx/services/spotify"
)

// Example demonstrating the new modular architecture
// This shows how V2 would handle multiple lyrics sources

func exampleMultiSourceSetup() {
	// Create the multi-source manager
	manager := sources.NewManager()

	// Add sources in priority order (highest priority first)

	// 1. Spotify (if available) - highest priority for accuracy
	if spotifyClient, err := spotify.New("your_cookie_here"); err == nil {
		manager.AddSource("spotify", spotifyClient, sources.PriorityCritical)
	}

	// 2. LRCLib - high priority, free and reliable
	lrclibClient := lrclib.NewClient()
	manager.AddSource("lrclib", lrclibClient, sources.PriorityHigh)

	// 3. Future sources can be easily added:
	// manager.AddSource("genius", geniusClient, sources.PriorityMedium)
	// manager.AddSource("musixmatch", musixmatchClient, sources.PriorityMedium)
	// manager.AddSource("azlyrics", azlyricsClient, sources.PriorityLow)

	// Example usage
	lines, err := manager.GetLyrics("", "artist|track|album|duration")
	if err != nil {
		log.Printf("Failed to get lyrics: %v", err)
		return
	}

	fmt.Printf("Found %d lines of lyrics\n", len(lines))

	// Example: Disable a source dynamically
	manager.EnableSource("spotify", false)

	// Example: List available sources
	sources := manager.GetSources()
	fmt.Println("Available sources:")
	for _, source := range sources {
		status := "enabled"
		if !source.Enabled {
			status = "disabled"
		}
		fmt.Printf("- %s (priority: %d, %s)\n", source.Name, source.Priority, status)
	}
}

// Configuration structure that anticipates V2 TOML config
type V2Config struct {
	LyricsSources LyricsSourcesConfig `toml:"lyrics_sources"`
	Players       PlayersConfig       `toml:"players"`
}

type LyricsSourcesConfig struct {
	Enabled  []string          `toml:"enabled"`
	Priority map[string]int    `toml:"priority"`
	Timeout  map[string]string `toml:"timeout"`

	// Source-specific configurations
	LRCLib  LRCLibConfig  `toml:"lrclib"`
	Spotify SpotifyConfig `toml:"spotify"`
	Genius  GeniusConfig  `toml:"genius"`
}

type LRCLibConfig struct {
	Enabled         bool   `toml:"enabled"`
	UseCachedOnly   bool   `toml:"use_cached_only"`
	CustomUserAgent string `toml:"custom_user_agent"`
}

type SpotifyConfig struct {
	Enabled bool   `toml:"enabled"`
	Cookie  string `toml:"cookie"`
}

type GeniusConfig struct {
	Enabled bool   `toml:"enabled"`
	APIKey  string `toml:"api_key"`
}

type PlayersConfig struct {
	Default  string   `toml:"default"`
	Priority []string `toml:"priority"`
}

// Example V2 TOML configuration
const exampleV2Config = `
[lyrics_sources]
enabled = ["lrclib", "spotify", "genius"]
priority = { spotify = 100, lrclib = 80, genius = 60 }
timeout = { spotify = "5s", lrclib = "10s", genius = "8s" }

[lyrics_sources.lrclib]
enabled = true
use_cached_only = false
custom_user_agent = "sptlrx/2.0.0"

[lyrics_sources.spotify]
enabled = true
cookie = "your_spotify_cookie"

[lyrics_sources.genius]
enabled = false
api_key = ""

[players]
default = "mpris"
priority = ["spotify", "mpris", "mpd"]
`
