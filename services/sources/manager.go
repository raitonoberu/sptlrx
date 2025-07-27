package sources

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/raitonoberu/sptlrx/lyrics"
)

var (
	ErrNoSources        = errors.New("no lyrics sources configured")
	ErrAllSourcesFailed = errors.New("all lyrics sources failed")
)

// Priority defines the priority of a lyrics source
type Priority int

const (
	PriorityLow Priority = iota
	PriorityMedium
	PriorityHigh
	PriorityCritical
)

// SourceConfig represents a configured lyrics source
type SourceConfig struct {
	Name     string
	Provider lyrics.Provider
	Priority Priority
	Timeout  time.Duration
	Enabled  bool
}

// Manager handles multiple lyrics sources with fallback logic (V2-ready)
type Manager struct {
	sources []SourceConfig
	timeout time.Duration
}

// NewManager creates a new multi-source lyrics manager
func NewManager() *Manager {
	return &Manager{
		sources: make([]SourceConfig, 0),
		timeout: 30 * time.Second, // Default global timeout
	}
}

// AddSource adds a lyrics source to the manager
func (m *Manager) AddSource(name string, provider lyrics.Provider, priority Priority) {
	config := SourceConfig{
		Name:     name,
		Provider: provider,
		Priority: priority,
		Timeout:  10 * time.Second, // Per-source timeout
		Enabled:  true,
	}

	// Insert maintaining priority order (highest first)
	inserted := false
	for i, existing := range m.sources {
		if priority > existing.Priority {
			// Insert at position i
			m.sources = append(m.sources[:i], append([]SourceConfig{config}, m.sources[i:]...)...)
			inserted = true
			break
		}
	}

	if !inserted {
		m.sources = append(m.sources, config)
	}
}

// GetLyrics attempts to get lyrics from sources in priority order
func (m *Manager) GetLyrics(id, query string) ([]lyrics.Line, error) {
	if len(m.sources) == 0 {
		return nil, ErrNoSources
	}

	ctx, cancel := context.WithTimeout(context.Background(), m.timeout)
	defer cancel()

	var lastErr error

	// Try sources in priority order
	for _, source := range m.sources {
		if !source.Enabled {
			continue
		}

		lines, err := m.trySource(ctx, source, id, query)
		if err == nil && len(lines) > 0 {
			// Success! Return with source info
			return lines, nil
		}

		lastErr = fmt.Errorf("source %s failed: %w", source.Name, err)
	}

	// If we get here, all sources failed
	if lastErr != nil {
		return nil, fmt.Errorf("%w: %v", ErrAllSourcesFailed, lastErr)
	}

	return nil, ErrAllSourcesFailed
}

// trySource attempts to get lyrics from a single source with timeout
func (m *Manager) trySource(ctx context.Context, source SourceConfig, id, query string) ([]lyrics.Line, error) {
	// Create source-specific context with timeout
	sourceCtx, cancel := context.WithTimeout(ctx, source.Timeout)
	defer cancel()

	// Channel for result
	resultCh := make(chan lyricsResult, 1)

	// Execute in goroutine to handle timeout
	go func() {
		lines, err := source.Provider.Lyrics(id, query)
		resultCh <- lyricsResult{lines: lines, err: err}
	}()

	// Wait for result or timeout
	select {
	case result := <-resultCh:
		return result.lines, result.err
	case <-sourceCtx.Done():
		return nil, fmt.Errorf("source %s timed out", source.Name)
	}
}

// GetSources returns the list of configured sources
func (m *Manager) GetSources() []SourceConfig {
	return append([]SourceConfig(nil), m.sources...) // Return copy
}

// EnableSource enables/disables a specific source
func (m *Manager) EnableSource(name string, enabled bool) error {
	for i := range m.sources {
		if m.sources[i].Name == name {
			m.sources[i].Enabled = enabled
			return nil
		}
	}
	return fmt.Errorf("source %s not found", name)
}

// SetSourceTimeout sets timeout for a specific source
func (m *Manager) SetSourceTimeout(name string, timeout time.Duration) error {
	for i := range m.sources {
		if m.sources[i].Name == name {
			m.sources[i].Timeout = timeout
			return nil
		}
	}
	return fmt.Errorf("source %s not found", name)
}

type lyricsResult struct {
	lines []lyrics.Line
	err   error
}
