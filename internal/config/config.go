package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type Config struct {
	ClientID     string   `toml:"client_id"`
	ClientSecret string   `toml:"client_secret"`
	Scopes       []string `toml:"scopes"`
	RedirectURI  string   `toml:"redirect_uri"`
	GrantType    string   `toml:"grant_type"`
	ActiveTenant string   `toml:"active_tenant"`
	Defaults     Defaults `toml:"defaults"`
}

type Defaults struct {
	Output   string `toml:"output"`
	PageSize int    `toml:"page_size"`
	CacheTTL string `toml:"cache_ttl"` // e.g. "5m", "1h", "0" to disable
}

func ConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}
	return filepath.Join(home, ".config", "xero-cli"), nil
}

func EnsureConfigDir() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", fmt.Errorf("cannot create config directory: %w", err)
	}
	return dir, nil
}

func ConfigPath() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.toml"), nil
}

func TokenPath() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "tokens.json"), nil
}

// newDefaults returns a Config with default values (no env overlay).
func newDefaults() *Config {
	return &Config{
		RedirectURI: "http://localhost:8472/callback",
		Scopes: []string{
			"openid", "offline_access",
			"accounting.invoices", "accounting.payments", "accounting.banktransactions",
			"accounting.manualjournals", "accounting.settings", "accounting.contacts",
			"accounting.attachments",
			"accounting.reports.aged.read", "accounting.reports.balancesheet.read",
			"accounting.reports.banksummary.read", "accounting.reports.budgetsummary.read",
			"accounting.reports.executivesummary.read", "accounting.reports.profitandloss.read",
			"accounting.reports.trialbalance.read", "accounting.reports.taxreports.read",
			"accounting.reports.tenninetynine.read", "accounting.budgets.read",
		},
		Defaults: Defaults{
			Output:   "table",
			PageSize: 100,
		},
	}
}

// resolvePath resolves the config file path: uses the given path, or falls back to the default.
func resolvePath(path string) (string, error) {
	if path != "" {
		return path, nil
	}
	return ConfigPath()
}

// LoadFile loads config from the TOML file only, without env-var overlays.
// Use this when you intend to mutate and Save() — avoids baking env-var secrets into the file.
func LoadFile(path string) (*Config, error) {
	cfg := newDefaults()

	resolved, err := resolvePath(path)
	if err != nil {
		return cfg, nil
	}

	if _, err := os.Stat(resolved); err == nil {
		if _, err := toml.DecodeFile(resolved, cfg); err != nil {
			return nil, fmt.Errorf("cannot parse config file %s: %w", resolved, err)
		}
	}

	return cfg, nil
}

// Load loads config from the TOML file, then applies env-var overlays.
func Load(path string) (*Config, error) {
	cfg, err := LoadFile(path)
	if err != nil {
		return nil, err
	}

	// Env var overlay
	if v := os.Getenv("XERO_CLIENT_ID"); v != "" {
		cfg.ClientID = v
	}
	if v := os.Getenv("XERO_CLIENT_SECRET"); v != "" {
		cfg.ClientSecret = v
	}
	if v := os.Getenv("XERO_TENANT_ID"); v != "" {
		cfg.ActiveTenant = v
	}
	if v := os.Getenv("XERO_GRANT_TYPE"); v != "" {
		cfg.GrantType = v
	}
	if v := os.Getenv("XERO_CACHE_TTL"); v != "" {
		cfg.Defaults.CacheTTL = v
	}

	return cfg, nil
}

func (c *Config) Save() error {
	path, err := ConfigPath()
	if err != nil {
		return err
	}
	dir, err := EnsureConfigDir()
	if err != nil {
		return err
	}

	// Atomic write: write to temp file, then rename
	tmp, err := os.CreateTemp(dir, ".config-*.toml.tmp")
	if err != nil {
		return fmt.Errorf("cannot create temp config file: %w", err)
	}
	tmpPath := tmp.Name()

	if err := toml.NewEncoder(tmp).Encode(c); err != nil {
		tmp.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("cannot encode config: %w", err)
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpPath)
		return err
	}

	if err := os.Rename(tmpPath, path); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("cannot write config: %w", err)
	}
	return nil
}
