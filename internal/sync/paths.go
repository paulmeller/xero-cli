package sync

import (
	"path/filepath"
	"strings"
)

// TenantStateFile returns a per-tenant state file path.
// e.g. ".xero-sync-state.json" -> ".xero-sync-state-256a364b.json"
func TenantStateFile(base, tenantID string) string {
	short := tenantID
	if len(short) > 8 {
		short = short[:8]
	}
	ext := filepath.Ext(base)
	return strings.TrimSuffix(base, ext) + "-" + short + ext
}

// TenantOutputDir returns a per-tenant output directory.
// e.g. "./xero_data" -> "./xero_data/256a364b"
func TenantOutputDir(base, tenantID string) string {
	short := tenantID
	if len(short) > 8 {
		short = short[:8]
	}
	return filepath.Join(base, short)
}
