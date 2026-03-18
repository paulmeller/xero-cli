package cmd

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/paulmeller/xero-cli/internal/cmdutil"
	"github.com/paulmeller/xero-cli/internal/config"
)

func newConfigCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "View and edit CLI configuration",
	}

	cmd.AddCommand(newConfigShowCmd(f))
	cmd.AddCommand(newConfigSetCmd(f))
	cmd.AddCommand(newConfigPathCmd())

	return cmd
}

func newConfigShowCmd(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Display current configuration",
		Example: `  xero config show
  xero config show -o json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := f.Config()
			if err != nil {
				return err
			}

			conn := cfg.ActiveConn()
			connName := cfg.ActiveConnectionName()
			format := cmdutil.GetOutputFormat(cmd, f.IO)

			if format == "json" {
				out := configJSON(cfg, conn, connName)
				data, err := json.MarshalIndent(out, "", "  ")
				if err != nil {
					return err
				}
				fmt.Fprintln(f.IO.Out, string(data))
				return nil
			}

			// Plain key=value output
			fmt.Fprintf(f.IO.Out, "connection = %s\n", connName)
			fmt.Fprintf(f.IO.Out, "client_id = %s\n", conn.ClientID)
			fmt.Fprintf(f.IO.Out, "client_secret = %s\n", redactSecret(conn.ClientSecret))
			fmt.Fprintf(f.IO.Out, "grant_type = %s\n", conn.GrantType)
			fmt.Fprintf(f.IO.Out, "redirect_uri = %s\n", conn.RedirectURI)
			fmt.Fprintf(f.IO.Out, "active_tenant = %s\n", conn.ActiveTenant)
			fmt.Fprintf(f.IO.Out, "scopes = [%d configured]\n", len(conn.Scopes))
			fmt.Fprintf(f.IO.Out, "defaults.output = %s\n", cfg.Defaults.Output)
			fmt.Fprintf(f.IO.Out, "defaults.page_size = %d\n", cfg.Defaults.PageSize)
			fmt.Fprintf(f.IO.Out, "defaults.cache_ttl = %s\n", cfg.Defaults.CacheTTL)

			return nil
		},
	}
}

func newConfigSetCmd(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set a configuration value",
		Args:  cobra.ExactArgs(2),
		Example: `  xero config set defaults.output json
  xero config set defaults.page_size 50
  xero config set defaults.cache_ttl 10m
  xero config set active_tenant <tenant-id>`,
		RunE: func(cmd *cobra.Command, args []string) error {
			key, value := args[0], args[1]

			// Reject sensitive keys
			switch key {
			case "client_id", "client_secret", "scopes":
				return fmt.Errorf("%s: use XERO_%s env var or edit config file directly",
					key, strings.ToUpper(key))
			}

			// Load from file only (no env overlay) to avoid leaking env-var secrets
			configPath, _ := cmd.Root().PersistentFlags().GetString("config")
			connName, _ := cmd.Root().PersistentFlags().GetString("connection")
			cfg, err := config.LoadFileWithConnection(configPath, connName)
			if err != nil {
				return err
			}

			switch key {
			case "active_tenant":
				cfg.SetActiveTenant(value)
			case "grant_type":
				if cfg.ActiveConnectionName() != "default" && cfg.Connections != nil {
					if conn, ok := cfg.Connections[cfg.ActiveConnection]; ok {
						conn.GrantType = value
					}
				} else {
					cfg.GrantType = value
				}
			case "redirect_uri":
				if cfg.ActiveConnectionName() != "default" && cfg.Connections != nil {
					if conn, ok := cfg.Connections[cfg.ActiveConnection]; ok {
						conn.RedirectURI = value
					}
				} else {
					cfg.RedirectURI = value
				}
			case "defaults.output":
				valid := map[string]bool{"table": true, "json": true, "csv": true, "tsv": true}
				if !valid[value] {
					return fmt.Errorf("invalid output format %q: must be table, json, csv, or tsv", value)
				}
				cfg.Defaults.Output = value
			case "defaults.page_size":
				n, err := strconv.Atoi(value)
				if err != nil || n < 1 || n > 100 {
					return fmt.Errorf("invalid page_size %q: must be an integer between 1 and 100", value)
				}
				cfg.Defaults.PageSize = n
			case "defaults.cache_ttl":
				if _, err := time.ParseDuration(value); err != nil {
					return fmt.Errorf("invalid cache_ttl %q: must be a Go duration (e.g. 5m, 1h, 0s)", value)
				}
				cfg.Defaults.CacheTTL = value
			default:
				return fmt.Errorf("unknown config key %q", key)
			}

			if err := cfg.Save(); err != nil {
				return err
			}

			quiet, _ := cmd.Root().PersistentFlags().GetBool("quiet")
			if !quiet {
				fmt.Fprintf(f.IO.ErrOut, "Set %s = %s\n", key, value)
			}
			return nil
		},
	}
}

func newConfigPathCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "path",
		Short: "Print config file path",
		RunE: func(cmd *cobra.Command, args []string) error {
			path, err := config.ConfigPath()
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), path)
			return nil
		},
	}
}

func redactSecret(s string) string {
	if s == "" {
		return ""
	}
	if len(s) <= 8 {
		return "****"
	}
	return s[:4] + "****" + s[len(s)-4:]
}

// configJSON builds a JSON-safe representation with the secret redacted.
func configJSON(cfg *config.Config, conn *config.Connection, connName string) map[string]any {
	return map[string]any{
		"connection":    connName,
		"client_id":     conn.ClientID,
		"client_secret": redactSecret(conn.ClientSecret),
		"grant_type":    conn.GrantType,
		"redirect_uri":  conn.RedirectURI,
		"active_tenant": conn.ActiveTenant,
		"scopes":        conn.Scopes,
		"defaults": map[string]any{
			"output":    cfg.Defaults.Output,
			"page_size": cfg.Defaults.PageSize,
			"cache_ttl": cfg.Defaults.CacheTTL,
		},
	}
}
