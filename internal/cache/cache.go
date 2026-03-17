package cache

import (
	"os"
	"path/filepath"
	"time"

	"github.com/paulmeller/xero-cli/internal/sync"
)

// IsFresh checks if a stream's cache is fresh (within TTL).
// Returns the JSONL file path if fresh, empty string if stale/missing.
func IsFresh(stateFile, outputDir, streamName string, ttl time.Duration) (jsonlPath string, ok bool) {
	if ttl <= 0 {
		return "", false
	}

	state, err := sync.LoadState(stateFile)
	if err != nil {
		return "", false
	}

	ss, exists := state.Streams[streamName]
	if !exists || ss.LastSync.IsZero() {
		return "", false
	}

	if time.Since(ss.LastSync) > ttl {
		return "", false
	}

	path := filepath.Join(outputDir, streamName+".jsonl")
	if _, err := os.Stat(path); err != nil {
		return "", false
	}

	return path, true
}
