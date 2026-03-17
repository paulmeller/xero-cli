package cache

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestIsFresh_ZeroLastSync(t *testing.T) {
	tmpDir := t.TempDir()
	path, ok := IsFresh(time.Time{}, tmpDir, "invoices", 10*time.Minute)
	if ok {
		t.Error("IsFresh should return false when lastSync is zero")
	}
	if path != "" {
		t.Errorf("path = %q, want empty", path)
	}
}

func TestIsFresh_StaleCache(t *testing.T) {
	tmpDir := t.TempDir()
	jsonlPath := filepath.Join(tmpDir, "invoices.jsonl")
	if err := os.WriteFile(jsonlPath, []byte(`{"InvoiceID":"1"}`+"\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// TTL of 5 minutes, last sync was 1 hour ago -> stale
	staleTime := time.Now().Add(-1 * time.Hour)
	path, ok := IsFresh(staleTime, tmpDir, "invoices", 5*time.Minute)
	if ok {
		t.Error("IsFresh should return false when cache is stale")
	}
	if path != "" {
		t.Errorf("path = %q, want empty", path)
	}
}

func TestIsFresh_MissingJSONL(t *testing.T) {
	tmpDir := t.TempDir()

	// Fresh sync time but no JSONL file
	path, ok := IsFresh(time.Now(), tmpDir, "invoices", 10*time.Minute)
	if ok {
		t.Error("IsFresh should return false when JSONL file is missing")
	}
	if path != "" {
		t.Errorf("path = %q, want empty", path)
	}
}

func TestIsFresh_FreshCache(t *testing.T) {
	tmpDir := t.TempDir()
	jsonlPath := filepath.Join(tmpDir, "invoices.jsonl")
	if err := os.WriteFile(jsonlPath, []byte(`{"InvoiceID":"1"}`+"\n"), 0644); err != nil {
		t.Fatal(err)
	}

	path, ok := IsFresh(time.Now(), tmpDir, "invoices", 10*time.Minute)
	if !ok {
		t.Error("IsFresh should return true when cache is fresh and JSONL exists")
	}
	if path != jsonlPath {
		t.Errorf("path = %q, want %q", path, jsonlPath)
	}
}

func TestIsFresh_ZeroTTL(t *testing.T) {
	path, ok := IsFresh(time.Now(), "/tmp", "invoices", 0)
	if ok {
		t.Error("IsFresh should return false when TTL is 0")
	}
	if path != "" {
		t.Errorf("path = %q, want empty", path)
	}
}
