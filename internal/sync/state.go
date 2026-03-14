package sync

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type SyncState struct {
	Version  int                    `json:"version"`
	TenantID string                 `json:"tenant_id"`
	Streams  map[string]StreamState `json:"streams"`
}

type StreamState struct {
	CursorValue   string    `json:"cursor_value"`
	LastSync      time.Time `json:"last_sync"`
	RecordsSynced int       `json:"records_synced"`
}

func LoadState(path string) (*SyncState, error) {
	state := &SyncState{
		Version: 1,
		Streams: make(map[string]StreamState),
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return state, nil
		}
		return nil, fmt.Errorf("cannot read state file: %w", err)
	}

	if err := json.Unmarshal(data, state); err != nil {
		return nil, fmt.Errorf("cannot parse state file: %w", err)
	}

	if state.Streams == nil {
		state.Streams = make(map[string]StreamState)
	}

	return state, nil
}

func SaveState(path string, state *SyncState) error {
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	tmp, err := os.CreateTemp(dir, ".sync-state-*.tmp")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpPath)
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpPath)
		return err
	}

	if err := os.Rename(tmpPath, path); err != nil {
		os.Remove(tmpPath)
		return err
	}

	return nil
}
