package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/zalando/go-keyring"
	"golang.org/x/oauth2"

	"github.com/paulmeller/xero-cli/internal/config"
)

const (
	keyringService = "xero-cli"
)

func keyringUserFor(connectionName string) string {
	if connectionName == "" || connectionName == "default" {
		return "oauth-token"
	}
	return "oauth-token:" + connectionName
}

// PersistentTokenSource implements oauth2.TokenSource with disk persistence.
// It wraps an underlying token source and saves tokens atomically on refresh.
// It always reloads the latest token from storage before refreshing, because
// Xero refresh tokens are single-use — another CLI invocation may have already
// rotated it.
type PersistentTokenSource struct {
	mu             sync.Mutex
	underlying     oauth2.TokenSource
	oauthCfg       *oauth2.Config // kept so we can rebuild token source with latest refresh token
	cached         *oauth2.Token
	connectionName string
}

func NewPersistentTokenSource(underlying oauth2.TokenSource, connectionName string) *PersistentTokenSource {
	return &PersistentTokenSource{
		underlying:     underlying,
		connectionName: connectionName,
	}
}

// NewPersistentTokenSourceWithConfig creates a PersistentTokenSource that can
// rebuild its underlying token source using the latest refresh token from storage.
// This is critical for Xero's single-use refresh tokens.
func NewPersistentTokenSourceWithConfig(oauthCfg *oauth2.Config, initial *oauth2.Token, connectionName string) *PersistentTokenSource {
	return &PersistentTokenSource{
		underlying:     oauthCfg.TokenSource(context.Background(), initial),
		oauthCfg:       oauthCfg,
		cached:         initial,
		connectionName: connectionName,
	}
}

func (s *PersistentTokenSource) Token() (*oauth2.Token, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Always load the latest token from storage — another CLI invocation
	// may have refreshed it since we last ran, and Xero refresh tokens
	// are single-use (the old one is invalidated on each refresh).
	stored, _ := LoadToken(s.connectionName)
	if stored != nil {
		s.cached = stored
	}

	// If cached token is valid, return it
	if s.cached != nil && s.cached.Valid() {
		return s.cached, nil
	}

	// Need to refresh — rebuild the token source with the latest refresh token
	// so we don't use a stale single-use refresh token.
	if s.cached == nil || s.cached.RefreshToken == "" {
		if s.underlying == nil {
			return nil, fmt.Errorf("no valid token and no token source configured; run 'xero auth login'")
		}
	}

	if s.oauthCfg != nil && s.cached != nil && s.cached.RefreshToken != "" {
		// Create a fresh token source using the latest refresh token
		s.underlying = s.oauthCfg.TokenSource(context.Background(), s.cached)
	}

	if s.underlying == nil {
		return nil, fmt.Errorf("no valid token and no token source configured; run 'xero auth login'")
	}

	tok, err := s.underlying.Token()
	if err != nil {
		return nil, fmt.Errorf("token refresh failed: %w", err)
	}

	s.cached = tok
	if err := SaveToken(s.connectionName, tok); err != nil {
		// Log but don't fail - we have a valid token in memory
		fmt.Fprintf(os.Stderr, "warning: could not save token: %v\n", err)
	}

	return tok, nil
}

// LoadToken loads the OAuth token, trying keychain first then falling back to file.
func LoadToken(connectionName string) (*oauth2.Token, error) {
	user := keyringUserFor(connectionName)

	// Try keychain first
	data, err := keyring.Get(keyringService, user)
	if err == nil {
		var tok oauth2.Token
		if err := json.Unmarshal([]byte(data), &tok); err == nil {
			return &tok, nil
		}
	}

	// Fall back to file
	return loadTokenFromFile(connectionName)
}

// SaveToken saves the OAuth token to keychain with file fallback.
func SaveToken(connectionName string, tok *oauth2.Token) error {
	user := keyringUserFor(connectionName)

	data, err := json.Marshal(tok)
	if err != nil {
		return err
	}

	// Try keychain first
	if err := keyring.Set(keyringService, user, string(data)); err == nil {
		// Success — clean up legacy file if it exists
		if path, err := config.TokenPathFor(connectionName); err == nil {
			os.Remove(path)
		}
		return nil
	}

	// Fall back to file
	return saveTokenToFile(connectionName, tok)
}

// DeleteToken removes the OAuth token from both keychain and file.
func DeleteToken(connectionName string) error {
	user := keyringUserFor(connectionName)

	// Delete from keychain (ignore error if not found)
	_ = keyring.Delete(keyringService, user)

	// Also delete file if it exists
	path, err := config.TokenPathFor(connectionName)
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// MigrateTokenToKeychain moves a file-based token to the OS keychain.
// Returns true if migration occurred, false if no file token existed.
func MigrateTokenToKeychain(connectionName string) (bool, error) {
	tok, err := loadTokenFromFile(connectionName)
	if err != nil {
		return false, nil // no file token to migrate
	}

	user := keyringUserFor(connectionName)

	data, err := json.Marshal(tok)
	if err != nil {
		return false, err
	}

	if err := keyring.Set(keyringService, user, string(data)); err != nil {
		return false, fmt.Errorf("keychain not available: %w", err)
	}

	// Remove file after successful migration
	if path, err := config.TokenPathFor(connectionName); err == nil {
		os.Remove(path)
	}

	return true, nil
}

func loadTokenFromFile(connectionName string) (*oauth2.Token, error) {
	path, err := config.TokenPathFor(connectionName)
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var tok oauth2.Token
	if err := json.Unmarshal(data, &tok); err != nil {
		return nil, fmt.Errorf("cannot parse token file: %w", err)
	}
	return &tok, nil
}

func saveTokenToFile(connectionName string, tok *oauth2.Token) error {
	path, err := config.TokenPathFor(connectionName)
	if err != nil {
		return err
	}
	if _, err := config.EnsureConfigDir(); err != nil {
		return err
	}

	data, err := json.MarshalIndent(tok, "", "  ")
	if err != nil {
		return err
	}

	// Atomic write: write to temp file then rename
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".tokens-*.tmp")
	if err != nil {
		return fmt.Errorf("cannot create temp file: %w", err)
	}
	tmpPath := tmp.Name()

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpPath)
		return err
	}
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
		return err
	}
	return nil
}
