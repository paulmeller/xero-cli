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

func Load(path string) (*Config, error) {
	cfg := &Config{
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

	if path == "" {
		var err error
		path, err = ConfigPath()
		if err != nil {
			return cfg, nil // Return defaults if can't find config
		}
	}

	if _, err := os.Stat(path); err == nil {
		if _, err := toml.DecodeFile(path, cfg); err != nil {
			return nil, fmt.Errorf("cannot parse config file %s: %w", path, err)
		}
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

	return cfg, nil
}

func (c *Config) Save() error {
	path, err := ConfigPath()
	if err != nil {
		return err
	}
	if _, err := EnsureConfigDir(); err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("cannot write config: %w", err)
	}
	defer f.Close()
	return toml.NewEncoder(f).Encode(c)
}
