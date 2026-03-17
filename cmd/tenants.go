package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"
	"golang.org/x/oauth2"

	"github.com/paulmeller/xero-cli/internal/api"
	"github.com/paulmeller/xero-cli/internal/auth"
	"github.com/paulmeller/xero-cli/internal/cmdutil"
	"github.com/paulmeller/xero-cli/internal/config"
	"github.com/paulmeller/xero-cli/internal/output"
)

func newTenantsCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "tenants",
		Aliases: []string{"tenant"},
		Short:   "Manage connected organizations",
	}

	cmd.AddCommand(newTenantsListCmd(f))
	cmd.AddCommand(newTenantsSwitchCmd(f))
	cmd.AddCommand(newTenantsCurrentCmd(f))

	return cmd
}

func newTenantsListCmd(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List connected organizations",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := tenantsClient(cmd, f)
			if err != nil {
				return err
			}

			data, err := client.GetConnections(cmd.Context())
			if err != nil {
				return err
			}

			cfg, _ := f.Config()
			format := cmdutil.GetOutputFormat(cmd, f.IO)

			if format == "json" {
				formatter := f.Formatter("json")
				return formatter.FormatList(f.IO.Out, gjson.ParseBytes(data), nil)
			}

			columns := []output.Column{
				{Header: "ACTIVE", Path: "_active"},
				{Header: "TENANT ID", Path: "tenantId"},
				{Header: "NAME", Path: "tenantName"},
				{Header: "TYPE", Path: "tenantType"},
			}

			// Add active marker
			tenants := gjson.ParseBytes(data)
			var enriched []map[string]any
			tenants.ForEach(func(_, t gjson.Result) bool {
				m := map[string]any{}
				json.Unmarshal([]byte(t.Raw), &m)
				if cfg != nil && t.Get("tenantId").String() == cfg.ActiveTenant {
					m["_active"] = "*"
				} else {
					m["_active"] = ""
				}
				enriched = append(enriched, m)
				return true
			})

			enrichedJSON, _ := json.Marshal(enriched)
			formatter := f.Formatter(format)
			return formatter.FormatList(f.IO.Out, gjson.ParseBytes(enrichedJSON), columns)
		},
	}
}

func newTenantsSwitchCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "switch [tenant-id]",
		Short: "Set the active tenant",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var tenantID string

			if len(args) > 0 {
				tenantID = args[0]
			} else {
				// Use --name to fuzzy match
				name, _ := cmd.Flags().GetString("name")
				if name == "" {
					return fmt.Errorf("provide a tenant ID or use --name to search")
				}

				client, err := tenantsClient(cmd, f)
				if err != nil {
					return err
				}

				data, err := client.GetConnections(cmd.Context())
				if err != nil {
					return err
				}

				tenants := gjson.ParseBytes(data)
				nameLower := strings.ToLower(name)
				var matches []gjson.Result

				tenants.ForEach(func(_, t gjson.Result) bool {
					tName := strings.ToLower(t.Get("tenantName").String())
					if strings.Contains(tName, nameLower) {
						matches = append(matches, t)
					}
					return true
				})

				if len(matches) == 0 {
					return fmt.Errorf("no tenant matching %q found", name)
				}
				if len(matches) > 1 {
					fmt.Fprintf(f.IO.ErrOut, "Multiple tenants match %q:\n", name)
					for _, m := range matches {
						fmt.Fprintf(f.IO.ErrOut, "  %s  %s\n", m.Get("tenantId").String(), m.Get("tenantName").String())
					}
					return fmt.Errorf("use tenant ID to be specific")
				}

				tenantID = matches[0].Get("tenantId").String()
				fmt.Fprintf(f.IO.ErrOut, "Matched: %s\n", matches[0].Get("tenantName").String())
			}

			// Use LoadFile to avoid baking env-var secrets into the config file
			fileCfg, _ := config.LoadFile("")
			fileCfg.ActiveTenant = tenantID
			if err := fileCfg.Save(); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}

			fmt.Fprintf(f.IO.ErrOut, "Active tenant set to: %s\n", tenantID)
			return nil
		},
	}

	cmd.Flags().String("name", "", "Fuzzy match tenant by organization name")

	return cmd
}

func newTenantsCurrentCmd(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "current",
		Short: "Show the active tenant",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := f.Config()
			if err != nil {
				return err
			}

			if cfg.ActiveTenant == "" {
				fmt.Fprintf(f.IO.ErrOut, "No active tenant set. Run 'xero tenants switch <id>'.\n")
				return nil
			}

			format := cmdutil.GetOutputFormat(cmd, f.IO)
			if format == "json" {
				data, _ := json.MarshalIndent(map[string]string{"tenant_id": cfg.ActiveTenant}, "", "  ")
				fmt.Fprintln(f.IO.Out, string(data))
				return nil
			}

			fmt.Fprintln(f.IO.Out, cfg.ActiveTenant)
			return nil
		},
	}
}

// tenantsClient creates a client suitable for the /connections endpoint.
// It doesn't require a tenant ID to be set.
func tenantsClient(cmd *cobra.Command, f *cmdutil.Factory) (*api.Client, error) {
	cfg, err := f.Config()
	if err != nil {
		return nil, err
	}

	tok, err := auth.LoadToken()
	if err != nil {
		return nil, fmt.Errorf("not authenticated; run 'xero auth login'")
	}

	oauthCfg := auth.OAuthConfig(cfg)
	ts := oauthCfg.TokenSource(cmd.Context(), tok)
	httpClient := oauth2.NewClient(cmd.Context(), ts)

	return api.NewClient(httpClient, "", false, false, io.Discard), nil
}
