package main

import (
	"fmt"

	"github.com/raitonoberu/sptlrx/services/hosted"
	"github.com/raitonoberu/sptlrx/services/lrclib"
	"github.com/raitonoberu/sptlrx/services/sources"
)

// Example demonstrating the multi-source manager usage
func main() {
	fmt.Println("=== sptlrx Multi-Source Manager Example ===")
	fmt.Println()

	// Create a new manager
	manager := sources.NewManager()

	// Add LRCLib source with priority 10
	lrclibProvider := lrclib.New()
	manager.AddSource(sources.SourceLRCLib, lrclibProvider, 10, true)
	fmt.Println("✓ Added LRCLib source (priority 10)")

	// Add hosted source with priority 20 (lower priority)
	hostedProvider := hosted.New("lyricsapi.vercel.app")
	manager.AddSource(sources.SourceHosted, hostedProvider, 20, true)
	fmt.Println("✓ Added Hosted API source (priority 20)")

	// Test with a known song
	testQueries := []string{
		"Rick Astley - Never Gonna Give You Up",
		"Queen - Bohemian Rhapsody",
		"Adele - Hello",
	}

	for i, query := range testQueries {
		fmt.Printf("\n--- Test %d: %s ---\n", i+1, query)

		lines, err := manager.Lyrics("", query)
		if err != nil {
			fmt.Printf("❌ Error: %v\n", err)
			continue
		}

		if len(lines) == 0 {
			fmt.Println("❌ No lyrics found")
			continue
		}

		fmt.Printf("✓ Found %d lyrics lines\n", len(lines))

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
	}

	// Show source statistics
	fmt.Println("\n=== Source Statistics ===")
	stats := manager.GetSourceStats()

	for sourceType, stat := range stats {
		fmt.Printf("%s:\n", sourceType)
		fmt.Printf("  Available: %v\n", stat.Available)
		fmt.Printf("  Success Rate: %.2f%%\n", stat.SuccessRate*100)
		fmt.Printf("  Avg Latency: %v\n", stat.AvgLatency)
		fmt.Println()
	}

	fmt.Println("=== Example Complete ===")
}
