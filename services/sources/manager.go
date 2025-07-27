package sources

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/raitonoberu/sptlrx/lyrics"
)

// SourceType represents the type of lyrics source
type SourceType string

const (
	SourceLocal      SourceType = "local"
	SourceLRCLib     SourceType = "lrclib"
	SourceGenius     SourceType = "genius"
	SourceMusixMatch SourceType = "musixmatch"
	SourceSpotify    SourceType = "spotify"
	SourceHosted     SourceType = "hosted"
)

// SourceHealth tracks the health status of a lyrics source
type SourceHealth struct {
	LastSuccess  time.Time
	LastFailure  time.Time
	SuccessCount int
	FailureCount int
	AvgLatency   time.Duration
	IsAvailable  bool
	mutex        sync.RWMutex
}

// UpdateSuccess records a successful lyrics retrieval
func (h *SourceHealth) UpdateSuccess(latency time.Duration) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	h.LastSuccess = time.Now()
	h.SuccessCount++
	h.IsAvailable = true

	// Calculate rolling average latency
	if h.AvgLatency == 0 {
		h.AvgLatency = latency
	} else {
		h.AvgLatency = (h.AvgLatency + latency) / 2
	}
}

// UpdateFailure records a failed lyrics retrieval
func (h *SourceHealth) UpdateFailure() {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	h.LastFailure = time.Now()
	h.FailureCount++

	// Mark as unavailable if too many consecutive failures
	if h.FailureCount > 3 && h.LastFailure.After(h.LastSuccess) {
		h.IsAvailable = false
	}
}

// GetHealth returns thread-safe health information
func (h *SourceHealth) GetHealth() (bool, time.Duration, float64) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	total := h.SuccessCount + h.FailureCount
	successRate := float64(0)
	if total > 0 {
		successRate = float64(h.SuccessCount) / float64(total)
	}

	return h.IsAvailable, h.AvgLatency, successRate
}

// SourceConfig defines configuration for a lyrics source
type SourceConfig struct {
	Type     SourceType
	Priority int
	Enabled  bool
	Timeout  time.Duration
	Provider lyrics.Provider
}

// Manager orchestrates multiple lyrics sources with intelligent fallback
type Manager struct {
	sources []SourceConfig
	health  map[SourceType]*SourceHealth
	cache   map[string]CacheEntry
	mutex   sync.RWMutex
}

// CacheEntry represents a cached lyrics result
type CacheEntry struct {
	Lines     []lyrics.Line
	Source    SourceType
	Timestamp time.Time
	TTL       time.Duration
}

// IsValid checks if the cache entry is still valid
func (e *CacheEntry) IsValid() bool {
	return time.Since(e.Timestamp) < e.TTL
}

// NewManager creates a new multi-source lyrics manager
func NewManager() *Manager {
	return &Manager{
		sources: make([]SourceConfig, 0),
		health:  make(map[SourceType]*SourceHealth),
		cache:   make(map[string]CacheEntry),
	}
}

// AddSource registers a new lyrics source with the manager
func (m *Manager) AddSource(sourceType SourceType, provider lyrics.Provider, priority int, enabled bool) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	config := SourceConfig{
		Type:     sourceType,
		Priority: priority,
		Enabled:  enabled,
		Timeout:  10 * time.Second,
		Provider: provider,
	}

	m.sources = append(m.sources, config)
	m.health[sourceType] = &SourceHealth{IsAvailable: true}

	// Sort sources by priority (lower number = higher priority)
	for i := len(m.sources) - 1; i > 0; i-- {
		if m.sources[i].Priority < m.sources[i-1].Priority {
			m.sources[i], m.sources[i-1] = m.sources[i-1], m.sources[i]
		} else {
			break
		}
	}
}

// Lyrics implements lyrics.Provider interface with multi-source intelligence
func (m *Manager) Lyrics(id, query string) ([]lyrics.Line, error) {
	// Check cache first
	if lines := m.getFromCache(query); lines != nil {
		return lines, nil
	}

	var lastErr error

	m.mutex.RLock()
	sources := make([]SourceConfig, len(m.sources))
	copy(sources, m.sources)
	m.mutex.RUnlock()

	// Try each source in priority order
	for _, source := range sources {
		if !source.Enabled {
			continue
		}

		health := m.health[source.Type]
		available, _, _ := health.GetHealth()
		if !available {
			continue
		}

		lines, err := m.trySource(source, id, query)
		if err != nil {
			lastErr = err
			health.UpdateFailure()
			continue
		}

		if len(lines) > 0 {
			// Cache successful result
			m.cacheResult(query, lines, source.Type)
			return lines, nil
		}
	}

	if lastErr != nil {
		return nil, fmt.Errorf("all sources failed, last error: %w", lastErr)
	}

	return nil, fmt.Errorf("no lyrics found from any source")
}

// trySource attempts to get lyrics from a specific source with timeout
func (m *Manager) trySource(source SourceConfig, id, query string) ([]lyrics.Line, error) {
	start := time.Now()

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), source.Timeout)
	defer cancel()

	// Channel to receive result
	type result struct {
		lines []lyrics.Line
		err   error
	}
	resultChan := make(chan result, 1)

	// Execute provider call in goroutine
	go func() {
		lines, err := source.Provider.Lyrics(id, query)
		resultChan <- result{lines: lines, err: err}
	}()

	// Wait for result or timeout
	select {
	case res := <-resultChan:
		latency := time.Since(start)
		if res.err == nil && len(res.lines) > 0 {
			m.health[source.Type].UpdateSuccess(latency)
		}
		return res.lines, res.err

	case <-ctx.Done():
		return nil, fmt.Errorf("source %s timed out after %v", source.Type, source.Timeout)
	}
}

// getFromCache retrieves lyrics from cache if available and valid
func (m *Manager) getFromCache(query string) []lyrics.Line {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	entry, exists := m.cache[query]
	if !exists || !entry.IsValid() {
		return nil
	}

	return entry.Lines
}

// cacheResult stores lyrics in cache with appropriate TTL
func (m *Manager) cacheResult(query string, lines []lyrics.Line, source SourceType) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Set TTL based on source type
	var ttl time.Duration
	switch source {
	case SourceLocal:
		ttl = 24 * time.Hour // Local files rarely change
	case SourceLRCLib:
		ttl = 12 * time.Hour // Community database updates occasionally
	case SourceGenius:
		ttl = 8 * time.Hour // Genius lyrics are fairly stable
	case SourceMusixMatch:
		ttl = 8 * time.Hour // MusixMatch lyrics are fairly stable
	case SourceSpotify:
		ttl = 6 * time.Hour // Spotify updates moderately
	case SourceHosted:
		ttl = 3 * time.Hour // External APIs might change more often
	default:
		ttl = 1 * time.Hour
	}

	m.cache[query] = CacheEntry{
		Lines:     lines,
		Source:    source,
		Timestamp: time.Now(),
		TTL:       ttl,
	}
}

// GetSourceStats returns statistics for all registered sources
func (m *Manager) GetSourceStats() map[SourceType]SourceStats {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	stats := make(map[SourceType]SourceStats)

	for sourceType, health := range m.health {
		available, avgLatency, successRate := health.GetHealth()

		stats[sourceType] = SourceStats{
			Type:        sourceType,
			Available:   available,
			SuccessRate: successRate,
			AvgLatency:  avgLatency,
		}
	}

	return stats
}

// SourceStats provides statistics about a lyrics source
type SourceStats struct {
	Type        SourceType
	Available   bool
	SuccessRate float64
	AvgLatency  time.Duration
}

// ClearCache removes all cached entries
func (m *Manager) ClearCache() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.cache = make(map[string]CacheEntry)
}

// RemoveSource removes a source from the manager
func (m *Manager) RemoveSource(sourceType SourceType) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Remove from sources slice
	for i, source := range m.sources {
		if source.Type == sourceType {
			m.sources = append(m.sources[:i], m.sources[i+1:]...)
			break
		}
	}

	// Remove from health tracking
	delete(m.health, sourceType)
}
