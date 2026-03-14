package destination

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type DuckDBDestination struct {
	connStr string
	tmpDir  string
}

func NewDuckDBDestination(connStr string) *DuckDBDestination {
	return &DuckDBDestination{connStr: connStr}
}

func (d *DuckDBDestination) Init(ctx context.Context) error {
	if _, err := exec.LookPath("duckdb"); err != nil {
		return fmt.Errorf("duckdb binary not found on PATH; install from https://duckdb.org")
	}

	var err error
	d.tmpDir, err = os.MkdirTemp("", "xero-sync-*")
	if err != nil {
		return err
	}

	return nil
}

func (d *DuckDBDestination) Write(ctx context.Context, stream string, primaryKey string, records []json.RawMessage) (int, error) {
	if len(records) == 0 {
		return 0, nil
	}

	// Write records to a temp JSONL file
	tmpFile := filepath.Join(d.tmpDir, stream+".jsonl")
	f, err := os.Create(tmpFile)
	if err != nil {
		return 0, err
	}

	for _, rec := range records {
		fmt.Fprintf(f, "%s\n", rec)
	}
	f.Close()

	// Build SQL for upsert
	tableName := sanitizeTableName(stream)

	sql := fmt.Sprintf(`
CREATE TABLE IF NOT EXISTS %s AS SELECT * FROM read_json_auto('%s') LIMIT 0;
INSERT OR REPLACE INTO %s SELECT * FROM read_json_auto('%s');
`, tableName, tmpFile, tableName, tmpFile)

	cmd := exec.CommandContext(ctx, "duckdb", d.connStr, "-c", sql)
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return 0, fmt.Errorf("duckdb command failed: %w", err)
	}

	os.Remove(tmpFile)
	return len(records), nil
}

func (d *DuckDBDestination) Close() error {
	if d.tmpDir != "" {
		return os.RemoveAll(d.tmpDir)
	}
	return nil
}

func sanitizeTableName(name string) string {
	return strings.ReplaceAll(name, "-", "_")
}
