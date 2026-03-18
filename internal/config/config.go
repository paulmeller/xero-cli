package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Connection holds credentials and settings for a single Xero app connection.
type Connection struct {
	ClientID     string   `toml:"client_id"`
	ClientSecret string   `toml:"client_secret"`
	Scopes       []string `toml:"scopes,omitempty"`
	RedirectURI  string   `toml:"redirect_uri,omitempty"`
	GrantType    string   `toml:"grant_type,omitempty"`
	ActiveTenant string   `toml:"active_tenant,omitempty"`
}

type Config struct {
	// Legacy flat fields (backward compat — used as "default" connection)
	ClientID     string   `toml:"client_id"`
	ClientSecret string   `toml:"client_secret"`
	Scopes       []string `toml:"scopes"`
	RedirectURI  string   `toml:"redirect_uri"`
	GrantType    string   `toml:"grant_type"`
	ActiveTenant string   `toml:"active_tenant"`

	// Multi-connection support
	ActiveConnection string                `toml:"active_connection,omitempty"`
	Connections      map[string]*Connection `toml:"connections,omitempty"`

	Defaults Defaults `toml:"defaults"`
}

type Defaults struct {
	Output   string `toml:"output"`
	PageSize int    `toml:"page_size"`
	CacheTTL string `toml:"cache_ttl"` // e.g. "5m", "1h", "0" to disable
}

// ActiveConnectionName returns the name of the active connection.
// Returns "default" when using legacy flat config fields.
func (c *Config) ActiveConnectionName() string {
	if c.ActiveConnection != "" {
		return c.ActiveConnection
	}
	return "default"
}

// ActiveConn returns the active connection's settings.
// If a named connection is active and exists in the map, that entry is returned.
// Otherwise, a Connection is synthesized from the legacy flat fields.
func (c *Config) ActiveConn() *Connection {
	if c.ActiveConnection != "" && c.Connections != nil {
		if conn, ok := c.Connections[c.ActiveConnection]; ok {
			// Fill in defaults from the top-level config for fields not set on the connection
			if len(conn.Scopes) == 0 {
				conn.Scopes = c.Scopes
			}
			if conn.RedirectURI == "" {
				conn.RedirectURI = c.RedirectURI
			}
			return conn
		}
	}
	// Synthesize from flat fields (backward compat)
	return &Connection{
		ClientID:     c.ClientID,
		ClientSecret: c.ClientSecret,
		Scopes:       c.Scopes,
		RedirectURI:  c.RedirectURI,
		GrantType:    c.GrantType,
		ActiveTenant: c.ActiveTenant,
	}
}

// SetActiveTenant sets the active tenant on the correct connection entry.
func (c *Config) SetActiveTenant(tenantID string) {
	if c.ActiveConnection != "" && c.Connections != nil {
		if conn, ok := c.Connections[c.ActiveConnection]; ok {
			conn.ActiveTenant = tenantID
			return
		}
	}
	c.ActiveTenant = tenantID
}

// SetActiveCredentials sets the client ID and secret on the correct connection entry.
func (c *Config) SetActiveCredentials(clientID, clientSecret string) {
	if c.ActiveConnection != "" && c.Connections != nil {
		if conn, ok := c.Connections[c.ActiveConnection]; ok {
			conn.ClientID = clientID
			conn.ClientSecret = clientSecret
			return
		}
	}
	c.ClientID = clientID
	c.ClientSecret = clientSecret
}

// SetConnection adds or updates a named connection.
func (c *Config) SetConnection(name string, conn *Connection) {
	if c.Connections == nil {
		c.Connections = make(map[string]*Connection)
	}
	c.Connections[name] = conn
}

// RemoveConnection removes a named connection from the map.
func (c *Config) RemoveConnection(name string) error {
	if c.Connections == nil {
		return fmt.Errorf("connection %q not found", name)
	}
	if _, ok := c.Connections[name]; !ok {
		return fmt.Errorf("connection %q not found", name)
	}
	delete(c.Connections, name)
	return nil
}

// ConnectionNames returns the names of all configured connections.
// Includes "default" if flat fields have a client_id set.
func (c *Config) ConnectionNames() []string {
	var names []string
	if c.ClientID != "" {
		names = append(names, "default")
	}
	for name := range c.Connections {
		names = append(names, name)
	}
	return names
}

// GetConnection returns a connection by name. "default" returns the synthesized flat-field connection.
func (c *Config) GetConnection(name string) (*Connection, bool) {
	if name == "default" || name == "" {
		if c.ClientID != "" {
			return &Connection{
				ClientID:     c.ClientID,
				ClientSecret: c.ClientSecret,
				Scopes:       c.Scopes,
				RedirectURI:  c.RedirectURI,
				GrantType:    c.GrantType,
				ActiveTenant: c.ActiveTenant,
			}, true
		}
		return nil, false
	}
	if c.Connections != nil {
		conn, ok := c.Connections[name]
		return conn, ok
	}
	return nil, false
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

// TokenPath returns the token file path for the default connection.
func TokenPath() (string, error) {
	return TokenPathFor("default")
}

// TokenPathFor returns the token file path for a named connection.
func TokenPathFor(connectionName string) (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	if connectionName == "" || connectionName == "default" {
		return filepath.Join(dir, "tokens.json"), nil
	}
	return filepath.Join(dir, "tokens-"+connectionName+".json"), nil
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

// LoadFileWithConnection loads config from file and sets the active connection override.
func LoadFileWithConnection(path, connectionOverride string) (*Config, error) {
	cfg, err := LoadFile(path)
	if err != nil {
		return nil, err
	}
	if connectionOverride != "" {
		cfg.ActiveConnection = connectionOverride
	}
	return cfg, nil
}

// Load loads config from the TOML file, then applies env-var overlays.
func Load(path string) (*Config, error) {
	return LoadWithConnection(path, "")
}

// LoadWithConnection loads config with an optional connection override, then applies env-var overlays.
func LoadWithConnection(path, connectionOverride string) (*Config, error) {
	cfg, err := LoadFile(path)
	if err != nil {
		return nil, err
	}

	// Apply connection override (flag takes precedence over env)
	if connectionOverride != "" {
		cfg.ActiveConnection = connectionOverride
	} else if v := os.Getenv("XERO_CONNECTION"); v != "" {
		cfg.ActiveConnection = v
	}

	// Env var overlays apply to the active connection
	conn := cfg.ActiveConn()
	if v := os.Getenv("XERO_CLIENT_ID"); v != "" {
		conn.ClientID = v
	}
	if v := os.Getenv("XERO_CLIENT_SECRET"); v != "" {
		conn.ClientSecret = v
	}
	if v := os.Getenv("XERO_TENANT_ID"); v != "" {
		conn.ActiveTenant = v
	}
	if v := os.Getenv("XERO_GRANT_TYPE"); v != "" {
		conn.GrantType = v
	}

	// Write back to the flat fields if using default connection (backward compat)
	if cfg.ActiveConnectionName() == "default" {
		cfg.ClientID = conn.ClientID
		cfg.ClientSecret = conn.ClientSecret
		cfg.ActiveTenant = conn.ActiveTenant
		cfg.GrantType = conn.GrantType
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
