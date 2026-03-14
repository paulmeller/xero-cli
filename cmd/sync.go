package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/paulmeller/xero-cli/internal/cmdutil"
	syncpkg "github.com/paulmeller/xero-cli/internal/sync"
	"github.com/paulmeller/xero-cli/internal/sync/destination"
)

func newSyncCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Sync Xero data locally (ELT)",
		Long:  "Sync data from Xero API to local JSONL files or DuckDB for offline querying.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSync(cmd, f)
		},
	}

	cmd.Flags().String("config-file", "sync.toml", "Path to sync config file")
	cmd.Flags().StringSlice("streams", nil, "Sync specific streams only (comma-separated)")
	cmd.Flags().Bool("full-refresh", false, "Ignore bookmarks, reload everything")

	cmd.AddCommand(newSyncInitCmd(f))
	cmd.AddCommand(newSyncStatusCmd(f))
	cmd.AddCommand(newSyncResetCmd(f))

	return cmd
}

func runSync(cmd *cobra.Command, f *cmdutil.Factory) error {
	configFile, _ := cmd.Flags().GetString("config-file")
	streamFilter, _ := cmd.Flags().GetStringSlice("streams")
	fullRefresh, _ := cmd.Flags().GetBool("full-refresh")
	dryRun, _ := cmd.Root().PersistentFlags().GetBool("dry-run")

	syncCfg, err := syncpkg.LoadSyncConfig(configFile)
	if err != nil {
		return err
	}

	state, err := syncpkg.LoadState(syncCfg.Sync.StateFile)
	if err != nil {
		return err
	}

	if fullRefresh {
		state.Streams = make(map[string]syncpkg.StreamState)
	}

	// Set tenant ID in state
	cfg, err := f.Config()
	if err != nil {
		return err
	}
	state.TenantID = cfg.ActiveTenant

	client, err := f.APIClient()
	if err != nil {
		return err
	}
	cmdutil.ApplyClientFlags(cmd, client, f)

	dest, err := createDestination(syncCfg.Destination)
	if err != nil {
		return err
	}

	engine := syncpkg.NewEngine(client, syncCfg, state, dest, f.IO.ErrOut, dryRun)

	if err := engine.Run(cmd.Context(), streamFilter); err != nil {
		return err
	}

	if !dryRun {
		if err := syncpkg.SaveState(syncCfg.Sync.StateFile, state); err != nil {
			return fmt.Errorf("failed to save sync state: %w", err)
		}
	}

	return nil
}

func newSyncInitCmd(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Generate default sync.toml",
		RunE: func(cmd *cobra.Command, args []string) error {
			path := "sync.toml"
			if _, err := os.Stat(path); err == nil {
				return fmt.Errorf("sync.toml already exists")
			}

			if err := os.WriteFile(path, []byte(syncpkg.DefaultSyncConfig()), 0644); err != nil {
				return err
			}

			fmt.Fprintf(f.IO.ErrOut, "Created %s\n", path)
			return nil
		},
	}
}

func newSyncStatusCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show last sync times and record counts",
		RunE: func(cmd *cobra.Command, args []string) error {
			configFile, _ := cmd.Parent().Flags().GetString("config-file")
			syncCfg, err := syncpkg.LoadSyncConfig(configFile)
			if err != nil {
				// If no config, try to load state directly
				syncCfg = &syncpkg.SyncConfig{
					Sync: syncpkg.SyncSettings{StateFile: ".xero-sync-state.json"},
				}
			}

			state, err := syncpkg.LoadState(syncCfg.Sync.StateFile)
			if err != nil {
				return err
			}

			format := cmdutil.GetOutputFormat(cmd, f.IO)
			if format == "json" {
				data, _ := json.MarshalIndent(state, "", "  ")
				fmt.Fprintln(f.IO.Out, string(data))
				return nil
			}

			if len(state.Streams) == 0 {
				fmt.Fprintln(f.IO.Out, "No sync history. Run 'xero sync' to start.")
				return nil
			}

			fmt.Fprintf(f.IO.Out, "%-25s %-25s %s\n", "STREAM", "LAST SYNC", "RECORDS")
			fmt.Fprintf(f.IO.Out, "%s\n", strings.Repeat("-", 65))

			for name, ss := range state.Streams {
				lastSync := "never"
				if !ss.LastSync.IsZero() {
					lastSync = ss.LastSync.Format(time.RFC3339)
				}
				fmt.Fprintf(f.IO.Out, "%-25s %-25s %d\n", name, lastSync, ss.RecordsSynced)
			}

			return nil
		},
	}

	return cmd
}

func newSyncResetCmd(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "reset [stream]",
		Short: "Clear incremental sync state",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			configFile, _ := cmd.Parent().Flags().GetString("config-file")
			stateFile := ".xero-sync-state.json"

			syncCfg, err := syncpkg.LoadSyncConfig(configFile)
			if err == nil {
				stateFile = syncCfg.Sync.StateFile
			}

			state, err := syncpkg.LoadState(stateFile)
			if err != nil {
				return err
			}

			if len(args) > 0 {
				delete(state.Streams, args[0])
				fmt.Fprintf(f.IO.ErrOut, "Reset sync state for %s\n", args[0])
			} else {
				state.Streams = make(map[string]syncpkg.StreamState)
				fmt.Fprintf(f.IO.ErrOut, "Reset all sync state\n")
			}

			return syncpkg.SaveState(stateFile, state)
		},
	}
}

func createDestination(cfg syncpkg.DestinationConfig) (destination.Destination, error) {
	switch cfg.Type {
	case "jsonl":
		return destination.NewJSONLDestination(cfg.OutputDir), nil
	case "duckdb":
		if cfg.ConnectionString == "" {
			return nil, fmt.Errorf("connection_string required for duckdb destination")
		}
		return destination.NewDuckDBDestination(cfg.ConnectionString), nil
	case "stdout":
		return destination.NewStdoutDestination(), nil
	default:
		return nil, fmt.Errorf("unknown destination type: %s", cfg.Type)
	}
}
