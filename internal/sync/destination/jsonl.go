package destination

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type JSONLDestination struct {
	outputDir string
	files     map[string]*os.File
}

func NewJSONLDestination(outputDir string) *JSONLDestination {
	return &JSONLDestination{
		outputDir: outputDir,
		files:     make(map[string]*os.File),
	}
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
