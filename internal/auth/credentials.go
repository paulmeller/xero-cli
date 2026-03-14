package auth

import (
	"context"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"

	"github.com/paulmeller/xero-cli/internal/config"
)

func ClientCredentialsTokenSource(ctx context.Context, cfg *config.Config) oauth2.TokenSource {
	ccCfg := &clientcredentials.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		TokenURL:     TokenURL,
		Scopes:       cfg.Scopes,
	}
	return ccCfg.TokenSource(ctx)
}
