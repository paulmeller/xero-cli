package cmdutil

import (
	"bufio"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// BuildListParams extracts common list parameters from command flags.
func BuildListParams(cmd *cobra.Command) url.Values {
	params := url.Values{}

	if v, _ := cmd.Flags().GetString("where"); v != "" {
		params.Set("where", v)
	}
	if v, _ := cmd.Flags().GetString("order"); v != "" {
		params.Set("order", v)
	}
	if v, _ := cmd.Flags().GetInt("page"); v > 0 {
		params.Set("page", fmt.Sprintf("%d", v))
	}
	if v, _ := cmd.Flags().GetInt("page-size"); v > 0 {
		if v > 100 {
			v = 100 // Xero API max
		}
		params.Set("pageSize", fmt.Sprintf("%d", v))
	}
	if v, _ := cmd.Flags().GetString("modified-since"); v != "" {
		params.Set("If-Modified-Since", v)
	}

	return params
}

// GetPageSize returns the effective page size from flag or config.
func GetPageSize(cmd *cobra.Command, f *Factory) int {
	if cmd.Root().PersistentFlags().Changed("page-size") {
		v, _ := cmd.Root().PersistentFlags().GetInt("page-size")
		if v > 100 {
			v = 100
		}
		return v
	}
	if cfg, err := f.Config(); err == nil && cfg.Defaults.PageSize > 0 {
		ps := cfg.Defaults.PageSize
		if ps > 100 {
			ps = 100
		}
		return ps
	}
	return 100
}

// ConfirmAction prompts the user for confirmation unless --force or --no-prompt is set.
// With --no-prompt and no --force, it returns false (fails safely for agents).
func ConfirmAction(ios *IOStreams, msg string, cmd *cobra.Command) bool {
	force, _ := cmd.Flags().GetBool("force")
	if force {
		return true
	}

	noPrompt, _ := cmd.Root().PersistentFlags().GetBool("no-prompt")
	if noPrompt || !ios.IsTTY {
		return false
	}

	fmt.Fprintf(ios.ErrOut, "%s [y/N]: ", msg)
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		answer := strings.TrimSpace(strings.ToLower(scanner.Text()))
		return answer == "y" || answer == "yes"
	}
	return false
}

// HasChangedFilterFlags returns true if any server-side filter flags have been set.
// Used to determine if a request can be served from cache.
func HasChangedFilterFlags(cmd *cobra.Command) bool {
	// Global filter flags
	globalFilters := []string{"where", "order", "page", "modified-since"}
	for _, name := range globalFilters {
		if cmd.Flags().Changed(name) {
			return true
		}
	}
	return false
}

// GetOutputFormat determines the output format from flags, config, or TTY detection.
func GetOutputFormat(cmd *cobra.Command, ios *IOStreams) string {
	format, _ := cmd.Root().PersistentFlags().GetString("output")
	if format != "" {
		return format
	}
	if !ios.IsTTY {
		return "json"
	}
	return "table"
}

// GetOutputFormatWithConfig determines the output format from flags, config defaults, or TTY detection.
func GetOutputFormatWithConfig(cmd *cobra.Command, f *Factory) string {
	// Explicit flag takes priority
	if cmd.Root().PersistentFlags().Changed("output") {
		format, _ := cmd.Root().PersistentFlags().GetString("output")
		return format
	}
	// Non-TTY always defaults to JSON regardless of config
	if !f.IO.IsTTY {
		return "json"
	}
	// Config default
	if cfg, err := f.Config(); err == nil && cfg.Defaults.Output != "" {
		return cfg.Defaults.Output
	}
	return "table"
}
