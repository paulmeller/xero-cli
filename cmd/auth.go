package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"
	"golang.org/x/oauth2"

	"github.com/paulmeller/xero-cli/internal/api"
	"github.com/paulmeller/xero-cli/internal/auth"
	"github.com/paulmeller/xero-cli/internal/cmdutil"
	"github.com/paulmeller/xero-cli/internal/output"
)

func newAuthCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Manage authentication",
	}

	cmd.AddCommand(newAuthLoginCmd(f))
	cmd.AddCommand(newAuthLogoutCmd(f))
	cmd.AddCommand(newAuthStatusCmd(f))
	cmd.AddCommand(newAuthRefreshCmd(f))
	cmd.AddCommand(newAuthMigrateKeychainCmd(f))

	return cmd
}

func newAuthLoginCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Authenticate with Xero",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := f.Config()
			if err != nil {
				return err
			}

			if cfg.ClientID == "" {
				return fmt.Errorf("client ID not configured; set XERO_CLIENT_ID or add client_id to config.toml")
			}

			headless, _ := cmd.Flags().GetBool("headless")
			ctx := cmd.Context()

			var tok *oauth2.Token
			if cfg.GrantType == "client_credentials" {
				ts := auth.ClientCredentialsTokenSource(ctx, cfg)
				tok, err = ts.Token()
				if err != nil {
					return fmt.Errorf("client credentials auth failed: %w", err)
				}
			} else if headless {
				readLine := func() (string, error) {
					scanner := bufio.NewScanner(os.Stdin)
					if scanner.Scan() {
						return scanner.Text(), nil
					}
					if err := scanner.Err(); err != nil {
						return "", err
					}
					return "", fmt.Errorf("no input")
				}
				tok, err = auth.LoginHeadless(ctx, cfg, f.IO.ErrOut, readLine)
				if err != nil {
					return err
				}
			} else {
				tok, err = auth.LoginInteractive(ctx, cfg, f.IO.ErrOut)
				if err != nil {
					return err
				}
			}

			if err := auth.SaveToken(tok); err != nil {
				return fmt.Errorf("failed to save token: %w", err)
			}

			fmt.Fprintf(f.IO.ErrOut, "Authentication successful!\n")

			// Fetch tenants and auto-select if only one.
			// Build a temporary client that doesn't require a tenant ID.
			httpClient := oauth2.NewClient(ctx, oauth2.StaticTokenSource(tok))
			client := api.NewClient(httpClient, "", false, false, io.Discard)

			data, err := client.GetConnections(ctx)
			if err != nil {
				fmt.Fprintf(f.IO.ErrOut, "Logged in. Run 'xero tenants list' to see your organizations.\n")
				return nil
			}

			tenants := gjson.ParseBytes(data)
			arr := tenants.Array()

			if len(arr) == 1 {
				tenantID := arr[0].Get("tenantId").String()
				tenantName := arr[0].Get("tenantName").String()
				cfg.ActiveTenant = tenantID
				if err := cfg.Save(); err != nil {
					fmt.Fprintf(f.IO.ErrOut, "Warning: could not save tenant to config: %v\n", err)
				}
				fmt.Fprintf(f.IO.ErrOut, "Active tenant: %s (%s)\n", tenantName, tenantID)
			} else if len(arr) > 1 {
				fmt.Fprintf(f.IO.ErrOut, "Found %d organizations. Run 'xero tenants switch' to select one.\n", len(arr))
			}

			return nil
		},
	}

	cmd.Flags().Bool("headless", false, "Print auth URL and accept callback URL paste (no browser)")

	return cmd
}

func newAuthLogoutCmd(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Remove stored authentication tokens",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := auth.DeleteToken(); err != nil {
				return err
			}
			fmt.Fprintf(f.IO.ErrOut, "Logged out.\n")
			return nil
		},
	}
}

func newAuthStatusCmd(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show authentication status",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := f.Config()
			if err != nil {
				return err
			}

			tok, err := auth.LoadToken()
			if err != nil {
				fmt.Fprintf(f.IO.ErrOut, "Not authenticated. Run 'xero auth login'.\n")
				return &cmdutil.SilentError{Code: cmdutil.ExitAuth}
			}

			format := cmdutil.GetOutputFormat(cmd, f.IO)
			if format == "json" {
				status := map[string]any{
					"authenticated": true,
					"tenant_id":     cfg.ActiveTenant,
					"token_expiry":  tok.Expiry.Format(time.RFC3339),
					"token_valid":   tok.Valid(),
					"grant_type":    cfg.GrantType,
				}
				data, _ := json.MarshalIndent(status, "", "  ")
				fmt.Fprintln(f.IO.Out, string(data))
				return nil
			}

			columns := []output.Column{
				{Header: "FIELD", Path: "field"},
				{Header: "VALUE", Path: "value"},
			}

			rows := []map[string]string{
				{"field": "Authenticated", "value": "Yes"},
				{"field": "Token Valid", "value": fmt.Sprintf("%v", tok.Valid())},
				{"field": "Token Expiry", "value": tok.Expiry.Format(time.RFC3339)},
				{"field": "Active Tenant", "value": cfg.ActiveTenant},
				{"field": "Grant Type", "value": cfg.GrantType},
			}

			rowsJSON, _ := json.Marshal(rows)
			items := gjson.ParseBytes(rowsJSON)

			formatter := f.Formatter(format)
			return formatter.FormatList(f.IO.Out, items, columns)
		},
	}
}

func newAuthMigrateKeychainCmd(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "migrate-keychain",
		Short: "Move stored token from file to OS keychain",
		RunE: func(cmd *cobra.Command, args []string) error {
			migrated, err := auth.MigrateTokenToKeychain()
			if err != nil {
				return err
			}
			if migrated {
				fmt.Fprintf(f.IO.ErrOut, "Token migrated to OS keychain. File-based token removed.\n")
			} else {
				fmt.Fprintf(f.IO.ErrOut, "No file-based token found. Nothing to migrate.\n")
			}
			return nil
		},
	}
}

func newAuthRefreshCmd(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "refresh",
		Short: "Force-refresh the access token",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := f.Config()
			if err != nil {
				return err
			}

			tok, err := auth.LoadToken()
			if err != nil {
				return fmt.Errorf("not authenticated; run 'xero auth login'")
			}

			oauthCfg := auth.OAuthConfig(cfg)
			ts := oauthCfg.TokenSource(cmd.Context(), tok)
			pts := auth.NewPersistentTokenSource(ts)

			newTok, err := pts.Token()
			if err != nil {
				return fmt.Errorf("token refresh failed: %w", err)
			}

			fmt.Fprintf(f.IO.ErrOut, "Token refreshed. Expires: %s\n", newTok.Expiry.Format(time.RFC3339))
			return nil
		},
	}
}

