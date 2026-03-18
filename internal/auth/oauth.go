package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"

	"github.com/pkg/browser"
	"golang.org/x/oauth2"

	"github.com/paulmeller/xero-cli/internal/config"
)

const (
	AuthURL  = "https://login.xero.com/identity/connect/authorize"
	TokenURL = "https://identity.xero.com/connect/token"
)

func OAuthConfig(conn *config.Connection) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     conn.ClientID,
		ClientSecret: conn.ClientSecret,
		Scopes:       conn.Scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:  AuthURL,
			TokenURL: TokenURL,
		},
		RedirectURL: conn.RedirectURI,
	}
}

func GenerateCodeVerifier() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func CodeChallenge(verifier string) string {
	h := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(h[:])
}

func LoginInteractive(ctx context.Context, conn *config.Connection, w io.Writer) (*oauth2.Token, error) {
	oauthCfg := OAuthConfig(conn)

	verifier, err := GenerateCodeVerifier()
	if err != nil {
		return nil, fmt.Errorf("failed to generate code verifier: %w", err)
	}
	challenge := CodeChallenge(verifier)

	state, err := generateState()
	if err != nil {
		return nil, fmt.Errorf("failed to generate state: %w", err)
	}

	srv := NewCallbackServer(conn.RedirectURI, state)
	if err := srv.Start(); err != nil {
		return nil, fmt.Errorf("failed to start callback server: %w", err)
	}
	defer srv.Close()

	authURL := oauthCfg.AuthCodeURL(
		state,
		oauth2.SetAuthURLParam("code_challenge", challenge),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
	)

	fmt.Fprintf(w, "Opening browser for authentication...\n")
	fmt.Fprintf(w, "If the browser doesn't open, visit:\n%s\n\n", authURL)

	_ = browser.OpenURL(authURL)

	fmt.Fprintf(w, "Waiting for callback...\n")
	code, err := srv.WaitForCode(ctx)
	if err != nil {
		return nil, err
	}

	tok, err := oauthCfg.Exchange(ctx, code,
		oauth2.SetAuthURLParam("code_verifier", verifier),
	)
	if err != nil {
		return nil, fmt.Errorf("token exchange failed: %w", err)
	}

	return tok, nil
}

func LoginHeadless(ctx context.Context, conn *config.Connection, w io.Writer, readLine func() (string, error)) (*oauth2.Token, error) {
	oauthCfg := OAuthConfig(conn)

	verifier, err := GenerateCodeVerifier()
	if err != nil {
		return nil, fmt.Errorf("failed to generate code verifier: %w", err)
	}
	challenge := CodeChallenge(verifier)

	state, err := generateState()
	if err != nil {
		return nil, fmt.Errorf("failed to generate state: %w", err)
	}

	authURL := oauthCfg.AuthCodeURL(
		state,
		oauth2.SetAuthURLParam("code_challenge", challenge),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
	)

	fmt.Fprintf(w, "Visit the following URL to authenticate:\n\n%s\n\n", authURL)
	fmt.Fprintf(w, "After authorizing, paste the full callback URL here:\n")

	callbackURL, err := readLine()
	if err != nil {
		return nil, fmt.Errorf("failed to read callback URL: %w", err)
	}

	code, err := ExtractCode(callbackURL, state)
	if err != nil {
		return nil, err
	}

	tok, err := oauthCfg.Exchange(ctx, code,
		oauth2.SetAuthURLParam("code_verifier", verifier),
	)
	if err != nil {
		return nil, fmt.Errorf("token exchange failed: %w", err)
	}

	return tok, nil
}

func generateState() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}
