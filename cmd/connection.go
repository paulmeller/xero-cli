package cmd

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"

	"github.com/paulmeller/xero-cli/internal/auth"
	"github.com/paulmeller/xero-cli/internal/cmdutil"
	"github.com/paulmeller/xero-cli/internal/config"
	"github.com/paulmeller/xero-cli/internal/output"
)

func newConnectionCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "connection",
		Aliases: []string{"conn"},
		Short:   "Manage named connection profiles",
		Long: `Manage multiple Xero app connections, each with its own
client ID, client secret, and OAuth token.

Use connections to switch between different Xero apps (e.g. dev vs production)
without reconfiguring credentials each time.`,
	}

	cmd.AddCommand(newConnectionListCmd(f))
	cmd.AddCommand(newConnectionAddCmd(f))
	cmd.AddCommand(newConnectionRemoveCmd(f))
	cmd.AddCommand(newConnectionSwitchCmd(f))
	cmd.AddCommand(newConnectionCurrentCmd(f))

	return cmd
}

func newConnectionListCmd(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all configured connections",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := f.Config()
			if err != nil {
				return err
			}

			activeConnName := cfg.ActiveConnectionName()
			format := cmdutil.GetOutputFormat(cmd, f.IO)

			type connInfo struct {
				Name     string
				ClientID string
				Grant    string
				Tenant   string
				Active   bool
			}

			var conns []connInfo

			// Add default connection if flat fields have a client_id
			if cfg.ClientID != "" {
				conns = append(conns, connInfo{
					Name:     "default",
					ClientID: cfg.ClientID,
					Grant:    cfg.GrantType,
					Tenant:   cfg.ActiveTenant,
					Active:   activeConnName == "default",
				})
			}

			// Add named connections
			if cfg.Connections != nil {
				names := make([]string, 0, len(cfg.Connections))
				for name := range cfg.Connections {
					names = append(names, name)
				}
				sort.Strings(names)

				for _, name := range names {
					conn := cfg.Connections[name]
					conns = append(conns, connInfo{
						Name:     name,
						ClientID: conn.ClientID,
						Grant:    conn.GrantType,
						Tenant:   conn.ActiveTenant,
						Active:   activeConnName == name,
					})
				}
			}

			if len(conns) == 0 {
				fmt.Fprintf(f.IO.ErrOut, "No connections configured. Run 'xero connection add <name>' or 'xero auth login'.\n")
				return nil
			}

			// Build JSON rows
			var rows []map[string]any
			for _, c := range conns {
				marker := ""
				if c.Active {
					marker = "*"
				}
				rows = append(rows, map[string]any{
					"_active":   marker,
					"name":      c.Name,
					"client_id": redactSecret(c.ClientID),
					"grant":     c.Grant,
					"tenant":    c.Tenant,
				})
			}

			rowsJSON, _ := json.Marshal(rows)

			if format == "json" {
				formatter := f.Formatter("json")
				return formatter.FormatList(f.IO.Out, gjson.ParseBytes(rowsJSON), nil)
			}

			columns := []output.Column{
				{Header: "ACTIVE", Path: "_active"},
				{Header: "NAME", Path: "name"},
				{Header: "CLIENT ID", Path: "client_id"},
				{Header: "GRANT TYPE", Path: "grant"},
				{Header: "ACTIVE TENANT", Path: "tenant"},
			}

			formatter := f.Formatter(format)
			return formatter.FormatList(f.IO.Out, gjson.ParseBytes(rowsJSON), columns)
		},
	}
}

func newConnectionAddCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add <name>",
		Short: "Add a new named connection",
		Args:  cobra.ExactArgs(1),
		Example: `  xero connection add production --client-id abc123 --client-secret secret456
  xero connection add staging --client-id def789`,
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			if name == "default" {
				return fmt.Errorf("cannot add a connection named \"default\"; the default connection uses the top-level config fields")
			}

			clientID, _ := cmd.Flags().GetString("client-id")
			clientSecret, _ := cmd.Flags().GetString("client-secret")
			grantType, _ := cmd.Flags().GetString("grant-type")
			switchTo, _ := cmd.Flags().GetBool("switch")

			noPrompt, _ := cmd.Root().PersistentFlags().GetBool("no-prompt")
			canPrompt := f.IO.IsTTY && !noPrompt

			if clientID == "" && canPrompt {
				var err error
				clientID, err = cmdutil.PromptString(f.IO, "Client ID: ")
				if err != nil {
					return err
				}
			}
			if clientID == "" {
				return fmt.Errorf("--client-id is required")
			}

			if clientSecret == "" && canPrompt {
				var err error
				clientSecret, err = cmdutil.PromptSecret(f.IO, "Client Secret (optional): ")
				if err != nil {
					return err
				}
			}

			// Load file config (no env overlay)
			configPath, _ := cmd.Root().PersistentFlags().GetString("config")
			cfg, err := config.LoadFile(configPath)
			if err != nil {
				return err
			}

			if cfg.Connections != nil {
				if _, exists := cfg.Connections[name]; exists {
					return fmt.Errorf("connection %q already exists; remove it first or choose a different name", name)
				}
			}

			conn := &config.Connection{
				ClientID:     clientID,
				ClientSecret: clientSecret,
				GrantType:    grantType,
			}
			cfg.SetConnection(name, conn)

			if switchTo {
				cfg.ActiveConnection = name
			}

			if err := cfg.Save(); err != nil {
				return err
			}

			fmt.Fprintf(f.IO.ErrOut, "Connection %q added.\n", name)
			if switchTo {
				fmt.Fprintf(f.IO.ErrOut, "Switched to connection %q.\n", name)
			} else {
				fmt.Fprintf(f.IO.ErrOut, "Run 'xero connection switch %s' to activate it, or use --switch.\n", name)
			}
			fmt.Fprintf(f.IO.ErrOut, "Run 'xero auth login -C %s' to authenticate.\n", name)
			return nil
		},
	}

	cmd.Flags().String("client-id", "", "Xero client ID")
	cmd.Flags().String("client-secret", "", "Xero client secret")
	cmd.Flags().String("grant-type", "", "OAuth grant type (leave empty for authorization code, or 'client_credentials')")
	cmd.Flags().Bool("switch", false, "Switch to this connection after adding")

	return cmd
}

func newConnectionRemoveCmd(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:     "remove <name>",
		Aliases: []string{"rm"},
		Short:   "Remove a named connection",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			if name == "default" {
				return fmt.Errorf("cannot remove the default connection; edit config.toml directly to clear the top-level fields")
			}

			configPath, _ := cmd.Root().PersistentFlags().GetString("config")
			cfg, err := config.LoadFile(configPath)
			if err != nil {
				return err
			}

			if cfg.ActiveConnection == name {
				return fmt.Errorf("cannot remove the active connection %q; switch to another connection first", name)
			}

			if err := cfg.RemoveConnection(name); err != nil {
				return err
			}

			if err := cfg.Save(); err != nil {
				return err
			}

			// Clean up the token for this connection
			_ = auth.DeleteToken(name)

			fmt.Fprintf(f.IO.ErrOut, "Connection %q removed.\n", name)
			return nil
		},
	}
}

func newConnectionSwitchCmd(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "switch <name>",
		Short: "Set the active connection",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			configPath, _ := cmd.Root().PersistentFlags().GetString("config")
			cfg, err := config.LoadFile(configPath)
			if err != nil {
				return err
			}

			// Validate that the connection exists
			if name == "default" {
				// Switch back to flat fields
				cfg.ActiveConnection = ""
			} else {
				if cfg.Connections == nil {
					return fmt.Errorf("connection %q not found", name)
				}
				if _, ok := cfg.Connections[name]; !ok {
					return fmt.Errorf("connection %q not found", name)
				}
				cfg.ActiveConnection = name
			}

			if err := cfg.Save(); err != nil {
				return err
			}

			fmt.Fprintf(f.IO.ErrOut, "Switched to connection %q.\n", name)
			return nil
		},
	}
}

func newConnectionCurrentCmd(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "current",
		Short: "Show the active connection",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := f.Config()
			if err != nil {
				return err
			}

			connName := cfg.ActiveConnectionName()
			conn := cfg.ActiveConn()

			format := cmdutil.GetOutputFormat(cmd, f.IO)
			if format == "json" {
				data, _ := json.MarshalIndent(map[string]string{
					"connection": connName,
					"client_id":  redactSecret(conn.ClientID),
					"tenant":     conn.ActiveTenant,
				}, "", "  ")
				fmt.Fprintln(f.IO.Out, string(data))
				return nil
			}

			fmt.Fprintf(f.IO.Out, "%s\n", connName)
			return nil
		},
	}
}
