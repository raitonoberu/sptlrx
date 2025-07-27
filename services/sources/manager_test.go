package sources

import (
	"fmt"
	"testing"
	"time"

	"github.com/raitonoberu/sptlrx/lyrics"
)

// MockProvider implements lyrics.Provider for testing
type MockProvider struct {
	name       string
	lyrics     []lyrics.Line
	delay      time.Duration
	shouldFail bool
}

func (m *MockProvider) Lyrics(id, query string) ([]lyrics.Line, error) {
	if m.delay > 0 {
		time.Sleep(m.delay)
	}

	if m.shouldFail {
		return nil, fmt.Errorf("mock provider %s failed", m.name)
	}

	return m.lyrics, nil
}

func TestNewManager(t *testing.T) {
	manager := NewManager()

	if manager == nil {
		t.Fatal("NewManager should return a non-nil manager")
	}

	if len(manager.sources) != 0 {
		t.Errorf("Expected 0 sources, got %d", len(manager.sources))
	}

	if len(manager.health) != 0 {
		t.Errorf("Expected 0 health entries, got %d", len(manager.health))
	}

	if len(manager.cache) != 0 {
		t.Errorf("Expected 0 cache entries, got %d", len(manager.cache))
	}
}

func TestAddSource(t *testing.T) {
	manager := NewManager()

	// Create mock providers
	localProvider := &MockProvider{
		name: "local",
		lyrics: []lyrics.Line{
			{Time: 0, Words: "Local lyrics"},
		},
	}

	lrclibProvider := &MockProvider{
		name: "lrclib",
		lyrics: []lyrics.Line{
			{Time: 1000, Words: "LRCLib lyrics"},
		},
	}

	// Add sources with different priorities
	manager.AddSource(SourceLRCLib, lrclibProvider, 20, true)
	manager.AddSource(SourceLocal, localProvider, 10, true)

	if len(manager.sources) != 2 {
		t.Errorf("Expected 2 sources, got %d", len(manager.sources))
	}

	// Check priority ordering (local should be first due to lower priority number)
	if manager.sources[0].Type != SourceLocal {
		t.Errorf("Expected first source to be local, got %s", manager.sources[0].Type)
	}

	if manager.sources[1].Type != SourceLRCLib {
		t.Errorf("Expected second source to be lrclib, got %s", manager.sources[1].Type)
	}
}

func TestLyricsSuccess(t *testing.T) {
	manager := NewManager()

	// Add a successful provider
	successProvider := &MockProvider{
		name: "success",
		lyrics: []lyrics.Line{
			{Time: 0, Words: "Test lyrics"},
			{Time: 1000, Words: "Second line"},
		},
	}

	manager.AddSource(SourceLocal, successProvider, 10, true)

	lines, err := manager.Lyrics("test-id", "test query")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(lines) != 2 {
		t.Errorf("Expected 2 lines, got %d", len(lines))
	}

	if lines[0].Words != "Test lyrics" {
		t.Errorf("Expected 'Test lyrics', got '%s'", lines[0].Words)
	}
}

func TestLyricsFallback(t *testing.T) {
	manager := NewManager()

	// Add failing provider with higher priority
	failProvider := &MockProvider{
		name:       "fail",
		shouldFail: true,
	}

	// Add successful provider with lower priority
	successProvider := &MockProvider{
		name: "success",
		lyrics: []lyrics.Line{
			{Time: 0, Words: "Fallback lyrics"},
		},
	}

	manager.AddSource(SourceLocal, failProvider, 10, true)
	manager.AddSource(SourceLRCLib, successProvider, 20, true)

	lines, err := manager.Lyrics("test-id", "test query")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(lines) != 1 {
		t.Errorf("Expected 1 line, got %d", len(lines))
	}

	if lines[0].Words != "Fallback lyrics" {
		t.Errorf("Expected 'Fallback lyrics', got '%s'", lines[0].Words)
	}
}

func TestLyricsAllFail(t *testing.T) {
	manager := NewManager()

	// Add only failing providers
	failProvider1 := &MockProvider{name: "fail1", shouldFail: true}
	failProvider2 := &MockProvider{name: "fail2", shouldFail: true}

	manager.AddSource(SourceLocal, failProvider1, 10, true)
	manager.AddSource(SourceLRCLib, failProvider2, 20, true)

	_, err := manager.Lyrics("test-id", "test query")
	if err == nil {
		t.Fatal("Expected error when all providers fail")
	}
}

func TestCaching(t *testing.T) {
	manager := NewManager()

	provider := &MockProvider{
		name: "cacheable",
		lyrics: []lyrics.Line{
			{Time: 0, Words: "Cached lyrics"},
		},
	}

	manager.AddSource(SourceLocal, provider, 10, true)

	// First call should hit the provider
	lines1, err := manager.Lyrics("test-id", "cache test")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Second call should hit the cache
	lines2, err := manager.Lyrics("test-id", "cache test")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(lines1) != len(lines2) {
		t.Errorf("Cache returned different results")
	}

	if lines1[0].Words != lines2[0].Words {
		t.Errorf("Cache returned different lyrics")
	}
}

func TestSourceHealth(t *testing.T) {
	health := &SourceHealth{IsAvailable: true}

	// Test successful update
	health.UpdateSuccess(100 * time.Millisecond)

	available, latency, successRate := health.GetHealth()
	if !available {
		t.Error("Expected source to be available after success")
	}

	if latency != 100*time.Millisecond {
		t.Errorf("Expected latency 100ms, got %v", latency)
	}

	if successRate != 1.0 {
		t.Errorf("Expected success rate 1.0, got %f", successRate)
	}

	// Test failure updates
	for i := 0; i < 4; i++ {
		health.UpdateFailure()
	}

	available, _, successRate = health.GetHealth()
	if available {
		t.Error("Expected source to be unavailable after multiple failures")
	}

	if successRate != 0.2 { // 1 success out of 5 total attempts
		t.Errorf("Expected success rate 0.2, got %f", successRate)
	}
}

func TestGetSourceStats(t *testing.T) {
	manager := NewManager()

	provider := &MockProvider{
		name:   "stats",
		lyrics: []lyrics.Line{{Time: 0, Words: "Stats test"}},
	}

	manager.AddSource(SourceLocal, provider, 10, true)

	// Make a successful call to generate stats
	_, err := manager.Lyrics("test-id", "stats test")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	stats := manager.GetSourceStats()

	localStats, exists := stats[SourceLocal]
	if !exists {
		t.Fatal("Expected stats for local source")
	}

	if !localStats.Available {
		t.Error("Expected local source to be available")
	}

	if localStats.SuccessRate != 1.0 {
		t.Errorf("Expected success rate 1.0, got %f", localStats.SuccessRate)
	}
}

func TestClearCache(t *testing.T) {
	manager := NewManager()

	provider := &MockProvider{
		name:   "clear",
		lyrics: []lyrics.Line{{Time: 0, Words: "Clear test"}},
	}

	manager.AddSource(SourceLocal, provider, 10, true)

	// Add something to cache
	_, err := manager.Lyrics("test-id", "clear test")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify cache has content
	if len(manager.cache) == 0 {
		t.Error("Expected cache to have content")
	}

	// Clear cache
	manager.ClearCache()

	// Verify cache is empty
	if len(manager.cache) != 0 {
		t.Errorf("Expected empty cache, got %d entries", len(manager.cache))
	}
}

func TestRemoveSource(t *testing.T) {
	manager := NewManager()

	provider := &MockProvider{
		name:   "remove",
		lyrics: []lyrics.Line{{Time: 0, Words: "Remove test"}},
	}

	manager.AddSource(SourceLocal, provider, 10, true)

	if len(manager.sources) != 1 {
		t.Errorf("Expected 1 source, got %d", len(manager.sources))
	}

	manager.RemoveSource(SourceLocal)

	if len(manager.sources) != 0 {
		t.Errorf("Expected 0 sources after removal, got %d", len(manager.sources))
	}

	if len(manager.health) != 0 {
		t.Errorf("Expected 0 health entries after removal, got %d", len(manager.health))
	}
}

func BenchmarkManagerLyrics(b *testing.B) {
	manager := NewManager()

	provider := &MockProvider{
		name: "bench",
		lyrics: []lyrics.Line{
			{Time: 0, Words: "Benchmark lyrics"},
			{Time: 1000, Words: "Second line"},
		},
	}

	manager.AddSource(SourceLocal, provider, 10, true)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := manager.Lyrics("bench-id", "benchmark query")
		if err != nil {
			b.Fatalf("Benchmark error: %v", err)
		}
	}
}
