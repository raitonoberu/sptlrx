package main

import (
	"fmt"

	"github.com/raitonoberu/sptlrx/services/genius"
	"github.com/raitonoberu/sptlrx/services/hosted"
	"github.com/raitonoberu/sptlrx/services/lrclib"
	"github.com/raitonoberu/sptlrx/services/musixmatch"
	"github.com/raitonoberu/sptlrx/services/sources"
)

// Example demonstrating all available lyrics sources with multi-source manager
func main() {
	fmt.Println("=== sptlrx Multi-Source Manager - All Sources Example ===")
	fmt.Println()

	// Create a new manager
	manager := sources.NewManager()

	// Add all available sources with different priorities
	fmt.Println("📚 Adding lyrics sources...")

	// Priority 10: LRCLib (free, community-driven, time-synced)
	lrclibProvider := lrclib.New()
	manager.AddSource(sources.SourceLRCLib, lrclibProvider, 10, true)
	fmt.Println("✓ LRCLib source added (priority 10) - Free community database with time-synced lyrics")

	// Priority 20: Genius (requires access token, rich lyrics content)
	// Note: In real usage, you would get this token from Genius API
	geniusProvider := genius.New("") // Empty token for demo - will show API key requirement
	manager.AddSource(sources.SourceGenius, geniusProvider, 20, true)
	fmt.Println("✓ Genius source added (priority 20) - Professional lyrics database (requires token)")

	// Priority 30: MusixMatch (requires API key, time-synced + regular lyrics)
	// Note: In real usage, you would get this API key from MusixMatch
	musixmatchProvider := musixmatch.New("") // Empty API key for demo
	manager.AddSource(sources.SourceMusixMatch, musixmatchProvider, 30, true)
	fmt.Println("✓ MusixMatch source added (priority 30) - Professional API with subtitles (requires API key)")

	// Priority 40: Hosted API (always available as fallback)
	hostedProvider := hosted.New("lyricsapi.vercel.app")
	manager.AddSource(sources.SourceHosted, hostedProvider, 40, true)
	fmt.Println("✓ Hosted API source added (priority 40) - Reliable fallback source")

	fmt.Println()

	// Test with various songs to demonstrate fallback behavior
	testQueries := []string{
		"Rick Astley - Never Gonna Give You Up",
		"Queen - Bohemian Rhapsody",
		"The Beatles - Hey Jude",
		"Adele - Hello",
		"Nonexistent Artist - Fake Song", // This should demonstrate fallback
	}

	fmt.Println("🎵 Testing lyrics retrieval with fallback system...")
	fmt.Println()

	for i, query := range testQueries {
		fmt.Printf("--- Test %d: %s ---\n", i+1, query)

		lines, err := manager.Lyrics("", query)
		if err != nil {
			fmt.Printf("❌ Error: %v\n", err)
			fmt.Println()
			continue
		}

		if len(lines) == 0 {
			fmt.Println("❌ No lyrics found from any source")
			fmt.Println()
			continue
		}

		fmt.Printf("✓ Found %d lyrics lines\n", len(lines))

		// Determine if lyrics are time-synced
		timeSynced := false
		for _, line := range lines {
			if line.Time > 0 {
				timeSynced = true
				break
			}
		}

		if timeSynced {
			fmt.Println("🕐 Time-synced lyrics available")
		} else {
			fmt.Println("📝 Plain text lyrics (no timing)")
		}

		// Show first few lines
		maxLines := 3
		if len(lines) < maxLines {
			maxLines = len(lines)
		}

		for j := 0; j < maxLines; j++ {
			if lines[j].Time > 0 {
				fmt.Printf("  [%02d:%02d.%03d] %s\n",
					lines[j].Time/60000,
					(lines[j].Time%60000)/1000,
					lines[j].Time%1000,
					lines[j].Words)
			} else {
				fmt.Printf("  %s\n", lines[j].Words)
			}
		}

		if len(lines) > maxLines {
			fmt.Printf("  ... and %d more lines\n", len(lines)-maxLines)
		}

		fmt.Println()
	}

	// Show comprehensive source statistics
	fmt.Println("=== 📊 Source Performance Statistics ===")
	stats := manager.GetSourceStats()

	sourceNames := map[sources.SourceType]string{
		sources.SourceLRCLib:     "LRCLib (Community)",
		sources.SourceGenius:     "Genius (Professional)",
		sources.SourceMusixMatch: "MusixMatch (API)",
		sources.SourceHosted:     "Hosted (Fallback)",
	}

	for sourceType, stat := range stats {
		name := sourceNames[sourceType]
		if name == "" {
			name = string(sourceType)
		}

		fmt.Printf("\n%s:\n", name)
		fmt.Printf("  🟢 Available: %v\n", stat.Available)
		fmt.Printf("  📈 Success Rate: %.1f%%\n", stat.SuccessRate*100)
		fmt.Printf("  ⏱️  Avg Latency: %v\n", stat.AvgLatency)

		// Add usage recommendations
		switch sourceType {
		case sources.SourceLRCLib:
			fmt.Printf("  💡 Best for: Popular songs with time-synced lyrics\n")
		case sources.SourceGenius:
			fmt.Printf("  💡 Best for: Detailed lyrics with annotations (needs token)\n")
		case sources.SourceMusixMatch:
			fmt.Printf("  💡 Best for: Professional subtitles and timing (needs API key)\n")
		case sources.SourceHosted:
			fmt.Printf("  💡 Best for: Reliable fallback when other sources fail\n")
		}
	}

	fmt.Println()
	fmt.Println("=== 🛠️ Configuration Recommendations ===")
	fmt.Println()
	fmt.Println("For optimal experience, configure in your sptlrx config.yaml:")
	fmt.Println()
	fmt.Println("```yaml")
	fmt.Println("# Enable/disable sources")
	fmt.Println("lrclib:")
	fmt.Println("  enabled: true  # Free, good coverage")
	fmt.Println()
	fmt.Println("genius:")
	fmt.Println("  enabled: true")
	fmt.Println("  accessToken: \"your_genius_token_here\"")
	fmt.Println()
	fmt.Println("musixmatch:")
	fmt.Println("  enabled: true")
	fmt.Println("  apiKey: \"your_musixmatch_key_here\"")
	fmt.Println()
	fmt.Println("sources:")
	fmt.Println("  cacheEnabled: true")
	fmt.Println("  cacheTTL: 3600")
	fmt.Println("  sourceTimeout: 10")
	fmt.Println("```")
	fmt.Println()
	fmt.Println("🔑 API Keys & Tokens:")
	fmt.Println("• Genius: Get free token at https://genius.com/api-clients")
	fmt.Println("• MusixMatch: Get API key at https://developer.musixmatch.com/")
	fmt.Println("• LRCLib: No authentication required (community-driven)")
	fmt.Println()
	fmt.Println("=== Example Complete ===")
}
