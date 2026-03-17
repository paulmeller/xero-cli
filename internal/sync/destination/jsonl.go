package destination

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type JSONLDestination struct {
	outputDir  string
	files      map[string]*os.File
	truncated  map[string]bool // tracks which streams have been truncated for full_refresh
}

func NewJSONLDestination(outputDir string) *JSONLDestination {
	return &JSONLDestination{
		outputDir: outputDir,
		files:     make(map[string]*os.File),
		truncated: make(map[string]bool),
	}
}

// TruncateStream truncates the JSONL file for a stream (used before full_refresh writes).
func (d *JSONLDestination) TruncateStream(stream string) error {
	if d.truncated[stream] {
		return nil
	}
	// Close existing file handle if open
	if f, ok := d.files[stream]; ok {
		f.Close()
		delete(d.files, stream)
	}
	path := filepath.Join(d.outputDir, stream+".jsonl")
	if err := os.Truncate(path, 0); err != nil && !os.IsNotExist(err) {
		return err
	}
	d.truncated[stream] = true
	return nil
}

func (d *JSONLDestination) Init(ctx context.Context) error {
	return os.MkdirAll(d.outputDir, 0755)
}

func (d *JSONLDestination) Write(ctx context.Context, stream string, primaryKey string, records []json.RawMessage) (int, error) {
	f, err := d.getFile(stream)
	if err != nil {
		return 0, err
	}

	written := 0
	for _, rec := range records {
		if _, err := f.Write(rec); err != nil {
			return written, err
		}
		if _, err := f.WriteString("\n"); err != nil {
			return written, err
		}
		written++
	}

	return written, nil
}

func (d *JSONLDestination) Close() error {
	var lastErr error
	for _, f := range d.files {
		if err := f.Close(); err != nil {
			lastErr = err
		}
	}
	return lastErr
}

func (d *JSONLDestination) getFile(stream string) (*os.File, error) {
	if f, ok := d.files[stream]; ok {
		return f, nil
	}

	path := filepath.Join(d.outputDir, stream+".jsonl")
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("cannot open %s: %w", path, err)
	}

	d.files[stream] = f
	return f, nil
}
