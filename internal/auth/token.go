package auth

import (
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
	keyringUser    = "oauth-token"
)

// PersistentTokenSource implements oauth2.TokenSource with disk persistence.
// It wraps an underlying token source and saves tokens atomically on refresh.
type PersistentTokenSource struct {
	mu         sync.Mutex
	underlying oauth2.TokenSource
	cached     *oauth2.Token
}

func NewPersistentTokenSource(underlying oauth2.TokenSource) *PersistentTokenSource {
	return &PersistentTokenSource{
		underlying: underlying,
	}
}

func (s *PersistentTokenSource) Token() (*oauth2.Token, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Try loading from storage if we have no cached token
	if s.cached == nil {
		t, _ := LoadToken()
		s.cached = t
	}

	// If cached token is valid, return it
	if s.cached != nil && s.cached.Valid() {
		return s.cached, nil
	}

	// Get a new token from the underlying source
	if s.underlying == nil {
		return nil, fmt.Errorf("no valid token and no token source configured; run 'xero auth login'")
	}

	tok, err := s.underlying.Token()
	if err != nil {
		return nil, fmt.Errorf("token refresh failed: %w", err)
	}

	s.cached = tok
	if err := SaveToken(tok); err != nil {
		// Log but don't fail - we have a valid token in memory
		fmt.Fprintf(os.Stderr, "warning: could not save token: %v\n", err)
	}

	return tok, nil
}

// LoadToken loads the OAuth token, trying keychain first then falling back to file.
func LoadToken() (*oauth2.Token, error) {
	// Try keychain first
	data, err := keyring.Get(keyringService, keyringUser)
	if err == nil {
		var tok oauth2.Token
		if err := json.Unmarshal([]byte(data), &tok); err == nil {
			return &tok, nil
		}
	}

	// Fall back to file
	return loadTokenFromFile()
}

// SaveToken saves the OAuth token to keychain with file fallback.
func SaveToken(tok *oauth2.Token) error {
	data, err := json.Marshal(tok)
	if err != nil {
		return err
	}

	// Try keychain first
	if err := keyring.Set(keyringService, keyringUser, string(data)); err == nil {
		// Success — clean up legacy file if it exists
		if path, err := config.TokenPath(); err == nil {
			os.Remove(path)
		}
		return nil
	}

	// Fall back to file
	return saveTokenToFile(tok)
}

// DeleteToken removes the OAuth token from both keychain and file.
func DeleteToken() error {
	// Delete from keychain (ignore error if not found)
	_ = keyring.Delete(keyringService, keyringUser)

	// Also delete file if it exists
	path, err := config.TokenPath()
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
func MigrateTokenToKeychain() (bool, error) {
	tok, err := loadTokenFromFile()
	if err != nil {
		return false, nil // no file token to migrate
	}

	data, err := json.Marshal(tok)
	if err != nil {
		return false, err
	}

	if err := keyring.Set(keyringService, keyringUser, string(data)); err != nil {
		return false, fmt.Errorf("keychain not available: %w", err)
	}

	// Remove file after successful migration
	if path, err := config.TokenPath(); err == nil {
		os.Remove(path)
	}

	return true, nil
}

func loadTokenFromFile() (*oauth2.Token, error) {
	path, err := config.TokenPath()
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

func saveTokenToFile(tok *oauth2.Token) error {
	path, err := config.TokenPath()
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
