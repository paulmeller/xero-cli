package auth

import (
	"context"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"

	"github.com/paulmeller/xero-cli/internal/config"
)

func ClientCredentialsTokenSource(ctx context.Context, conn *config.Connection) oauth2.TokenSource {
	ccCfg := &clientcredentials.Config{
		ClientID:     conn.ClientID,
		ClientSecret: conn.ClientSecret,
		TokenURL:     TokenURL,
		Scopes:       conn.Scopes,
	}
	return ccCfg.TokenSource(ctx)
}
