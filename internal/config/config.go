package config

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"

	"github.com/BurntSushi/toml"
)

var validConnectionName = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]{0,63}$`)

// ValidateConnectionName checks that a connection name is safe for use
// in file paths, keyring keys, and TOML map keys.
func ValidateConnectionName(name string) error {
	if name == "" {
		return fmt.Errorf("connection name cannot be empty")
	}
	if !validConnectionName.MatchString(name) {
		return fmt.Errorf("connection name %q is invalid: must be 1-64 alphanumeric characters, hyphens, or underscores, starting with an alphanumeric character", name)
	}
	return nil
}

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
	// Global defaults — connections inherit these when their own fields are empty.
	Scopes      []string `toml:"scopes"`
	RedirectURI string   `toml:"redirect_uri"`

	// Active connection name. Empty means "default".
	ActiveConnection string                `toml:"active_connection,omitempty"`
	Connections      map[string]*Connection `toml:"connections"`

	Defaults Defaults `toml:"defaults"`

	// Legacy flat fields — only used for migration from old config format.
	// Cleared after load; omitempty prevents them from being saved.
	ClientID     string `toml:"client_id,omitempty"`
	ClientSecret string `toml:"client_secret,omitempty"`
	GrantType    string `toml:"grant_type,omitempty"`
	ActiveTenant string `toml:"active_tenant,omitempty"`
}

type Defaults struct {
	Output   string `toml:"output"`
	PageSize int    `toml:"page_size"`
	CacheTTL string `toml:"cache_ttl"` // e.g. "5m", "1h", "0" to disable
}

// migrate moves legacy flat credential fields into connections["default"]
// and clears them so they won't be written back on Save.
func (c *Config) migrate() {
	if c.ClientID == "" {
		return
	}
	if c.Connections == nil {
		c.Connections = make(map[string]*Connection)
	}
	if _, exists := c.Connections["default"]; !exists {
		c.Connections["default"] = &Connection{
			ClientID:     c.ClientID,
			ClientSecret: c.ClientSecret,
			GrantType:    c.GrantType,
			ActiveTenant: c.ActiveTenant,
		}
	}
	c.ClientID = ""
	c.ClientSecret = ""
	c.GrantType = ""
	c.ActiveTenant = ""
}

// ActiveConnectionName returns the name of the active connection.
// Returns "default" when no explicit connection is set.
func (c *Config) ActiveConnectionName() string {
	if c.ActiveConnection != "" {
		return c.ActiveConnection
	}
	return "default"
}

// ActiveConn returns a copy of the active connection's settings.
// Global scopes and redirect_uri are filled in when the connection
// doesn't specify its own. Always returns a non-nil Connection.
func (c *Config) ActiveConn() *Connection {
	name := c.ActiveConnectionName()
	if c.Connections != nil {
		if conn, ok := c.Connections[name]; ok {
			cp := *conn
			if len(cp.Scopes) == 0 {
				cp.Scopes = append([]string(nil), c.Scopes...)
			}
			if cp.RedirectURI == "" {
				cp.RedirectURI = c.RedirectURI
			}
			return &cp
		}
	}
	// No connection found — return empty connection with global defaults.
	return &Connection{
		Scopes:      append([]string(nil), c.Scopes...),
		RedirectURI: c.RedirectURI,
	}
}

// SetActiveTenant sets the active tenant on the active connection,
// creating the connection entry if it doesn't exist.
func (c *Config) SetActiveTenant(tenantID string) {
	conn := c.ensureActiveConn()
	conn.ActiveTenant = tenantID
}

// SetActiveCredentials sets the client ID and secret on the active connection,
// creating the connection entry if it doesn't exist.
func (c *Config) SetActiveCredentials(clientID, clientSecret string) {
	conn := c.ensureActiveConn()
	conn.ClientID = clientID
	conn.ClientSecret = clientSecret
}

// SetActiveGrantType sets the grant type on the active connection,
// creating the connection entry if it doesn't exist.
func (c *Config) SetActiveGrantType(grantType string) {
	conn := c.ensureActiveConn()
	conn.GrantType = grantType
}

// SetActiveRedirectURI sets the redirect URI on the active connection,
// creating the connection entry if it doesn't exist.
func (c *Config) SetActiveRedirectURI(uri string) {
	conn := c.ensureActiveConn()
	conn.RedirectURI = uri
}

// ensureActiveConn returns the active connection entry, creating it if needed.
func (c *Config) ensureActiveConn() *Connection {
	name := c.ActiveConnectionName()
	if c.Connections == nil {
		c.Connections = make(map[string]*Connection)
	}
	conn, ok := c.Connections[name]
	if !ok {
		conn = &Connection{}
		c.Connections[name] = conn
	}
	return conn
}

// SetConnection adds or updates a named connection.
// Returns an error if the name is invalid.
func (c *Config) SetConnection(name string, conn *Connection) error {
	if err := ValidateConnectionName(name); err != nil {
		return err
	}
	if c.Connections == nil {
		c.Connections = make(map[string]*Connection)
	}
	c.Connections[name] = conn
	return nil
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

// ConnectionNames returns the sorted names of all configured connections.
func (c *Config) ConnectionNames() []string {
	names := make([]string, 0, len(c.Connections))
	for name := range c.Connections {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// GetConnection returns a copy of a connection by name.
func (c *Config) GetConnection(name string) (*Connection, bool) {
	if c.Connections != nil {
		if conn, ok := c.Connections[name]; ok {
			cp := *conn
			return &cp, true
		}
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
// Legacy flat-field configs are automatically migrated to the connections map.
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

	cfg.migrate()
	return cfg, nil
}

// LoadFileWithConnection loads config from file and sets the active connection override.
// Returns an error if the override names a connection that does not exist.
func LoadFileWithConnection(path, connectionOverride string) (*Config, error) {
	cfg, err := LoadFile(path)
	if err != nil {
		return nil, err
	}
	if connectionOverride != "" && connectionOverride != "default" {
		if cfg.Connections[connectionOverride] == nil {
			return nil, fmt.Errorf("connection %q not found", connectionOverride)
		}
	}
	if connectionOverride != "" {
		if connectionOverride == "default" {
			cfg.ActiveConnection = ""
		} else {
			cfg.ActiveConnection = connectionOverride
		}
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
	override := connectionOverride
	if override == "" {
		override = os.Getenv("XERO_CONNECTION")
	}
	if override != "" {
		// Validate that the named connection exists (env vars may still create it below)
		if override != "default" && override != "" {
			if cfg.Connections[override] == nil {
				return nil, fmt.Errorf("connection %q not found", override)
			}
		}
		if override == "default" {
			cfg.ActiveConnection = ""
		} else {
			cfg.ActiveConnection = override
		}
	}

	// Env var overlays — applied to the active connection in-memory.
	// This config should NOT be saved; use LoadFile for mutation paths.
	conn := cfg.ensureActiveConn()
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
	// Config may contain client_secret — restrict permissions.
	if err := tmp.Chmod(0600); err != nil {
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
		return fmt.Errorf("cannot write config: %w", err)
	}
	return nil
}
