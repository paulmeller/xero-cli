package sync

import (
	"path/filepath"
	"testing"
)

func TestTenantStateFile_Standard(t *testing.T) {
	got := TenantStateFile(".xero-sync-state.json", "256a364b-1234-5678-9abc-def012345678")
	want := ".xero-sync-state-256a364b.json"
	if got != want {
		t.Errorf("TenantStateFile = %q, want %q", got, want)
	}
}

func TestTenantStateFile_ShortTenantID(t *testing.T) {
	got := TenantStateFile(".xero-sync-state.json", "abc")
	want := ".xero-sync-state-abc.json"
	if got != want {
		t.Errorf("TenantStateFile = %q, want %q", got, want)
	}
}

func TestTenantStateFile_ExactlyEightChars(t *testing.T) {
	got := TenantStateFile(".xero-sync-state.json", "12345678")
	want := ".xero-sync-state-12345678.json"
	if got != want {
		t.Errorf("TenantStateFile = %q, want %q", got, want)
	}
}

func TestTenantStateFile_WithDirectory(t *testing.T) {
	got := TenantStateFile(filepath.Join("some", "dir", "state.json"), "256a364b-xxxx")
	want := filepath.Join("some", "dir", "state-256a364b.json")
	if got != want {
		t.Errorf("TenantStateFile = %q, want %q", got, want)
	}
}

func TestTenantOutputDir_Standard(t *testing.T) {
	got := TenantOutputDir("./xero_data", "256a364b-1234-5678-9abc-def012345678")
	want := filepath.Join("xero_data", "256a364b")
	if got != want {
		t.Errorf("TenantOutputDir = %q, want %q", got, want)
	}
}

func TestTenantOutputDir_ShortTenantID(t *testing.T) {
	got := TenantOutputDir("./xero_data", "abc")
	want := filepath.Join("xero_data", "abc")
	if got != want {
		t.Errorf("TenantOutputDir = %q, want %q", got, want)
	}
}

func TestTenantOutputDir_AbsolutePath(t *testing.T) {
	got := TenantOutputDir("/tmp/xero_data", "abcdef12-xxxx")
	want := filepath.Join("/tmp/xero_data", "abcdef12")
	if got != want {
		t.Errorf("TenantOutputDir = %q, want %q", got, want)
	}
}
