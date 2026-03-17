package cache

import (
	"os"
	"path/filepath"
	"time"
)

// IsFresh checks if a stream's cache is fresh (within TTL).
// lastSync is the time the stream was last synced.
// Returns the JSONL file path if fresh, empty string if stale/missing.
func IsFresh(lastSync time.Time, outputDir, streamName string, ttl time.Duration) (jsonlPath string, ok bool) {
	if ttl <= 0 {
		return "", false
	}

	if lastSync.IsZero() {
		return "", false
	}

	if time.Since(lastSync) > ttl {
		return "", false
	}

	path := filepath.Join(outputDir, streamName+".jsonl")
	if _, err := os.Stat(path); err != nil {
		return "", false
	}

	return path, true
}
