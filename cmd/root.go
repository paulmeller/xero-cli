package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/paulmeller/xero-cli/internal/cmdutil"
)

var Version = "dev"

func Execute() int {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	f := cmdutil.NewFactory()

	rootCmd := &cobra.Command{
		Use:           "xero",
		Short:         "Xero accounting CLI",
		Long: `A command-line interface for the Xero accounting API.

Enable shell completions: xero completion --help`,
		SilenceErrors: true,
		SilenceUsage:  true,
		Version:       Version,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			cmdutil.BindTokenFlag(cmd, f)
		},
	}

	// Global persistent flags
	rootCmd.PersistentFlags().StringP("output", "o", "", "Output format: table, json, csv, tsv")
	rootCmd.PersistentFlags().StringP("tenant", "t", "", "Xero tenant ID (overrides config)")
	rootCmd.PersistentFlags().BoolP("quiet", "q", false, "Suppress non-essential output")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Verbose output (print HTTP requests/responses)")
	rootCmd.PersistentFlags().Bool("no-color", false, "Disable color output")
	rootCmd.PersistentFlags().Bool("no-prompt", false, "Never prompt for confirmation; fail if confirmation required")
	rootCmd.PersistentFlags().Int("page", 0, "Page number for paginated results")
	rootCmd.PersistentFlags().Int("page-size", 100, "Items per page (max 100)")
	rootCmd.PersistentFlags().String("modified-since", "", "Only return items modified after this date (ISO 8601)")
	rootCmd.PersistentFlags().String("where", "", "Filter expression (Xero where clause)")
	rootCmd.PersistentFlags().String("order", "", "Sort expression")
	rootCmd.PersistentFlags().String("token", "", "Use external access token (bypasses stored token)")
	rootCmd.PersistentFlags().Bool("dry-run", false, "Print requests without sending them")
	rootCmd.PersistentFlags().String("config", "", "Path to config file")
	rootCmd.PersistentFlags().Int("timeout", 30, "Request timeout in seconds")
	rootCmd.PersistentFlags().Bool("live", false, "Bypass local cache, always fetch from API")
	rootCmd.PersistentFlags().Duration("cache-ttl", 5*time.Minute, "Cache freshness duration (0 to disable)")

	// Auto-enable --no-prompt when stdin is not a TTY
	if !f.IO.IsTTY {
		rootCmd.PersistentFlags().Set("no-prompt", "true")
	}

	// Register commands
	rootCmd.AddCommand(newAuthCmd(f))
	rootCmd.AddCommand(newTenantsCmd(f))
	rootCmd.AddCommand(newCompletionCmd())
	rootCmd.AddCommand(newInvoicesCmd(f))
	rootCmd.AddCommand(newContactsCmd(f))
	rootCmd.AddCommand(newPaymentsCmd(f))
	rootCmd.AddCommand(newAccountsCmd(f))
	rootCmd.AddCommand(newCreditNotesCmd(f))
	rootCmd.AddCommand(newBankTransactionsCmd(f))
	rootCmd.AddCommand(newPurchaseOrdersCmd(f))
	rootCmd.AddCommand(newItemsCmd(f))
	rootCmd.AddCommand(newManualJournalsCmd(f))
	rootCmd.AddCommand(newJournalsCmd(f))
	rootCmd.AddCommand(newQuotesCmd(f))
	rootCmd.AddCommand(newRepeatingInvoicesCmd(f))
	rootCmd.AddCommand(newBatchPaymentsCmd(f))
	rootCmd.AddCommand(newOverpaymentsCmd(f))
	rootCmd.AddCommand(newPrepaymentsCmd(f))
	rootCmd.AddCommand(newLinkedTransactionsCmd(f))
	rootCmd.AddCommand(newTaxRatesCmd(f))
	rootCmd.AddCommand(newCurrenciesCmd(f))
	rootCmd.AddCommand(newBrandingThemesCmd(f))
	rootCmd.AddCommand(newTrackingCmd(f))
	rootCmd.AddCommand(newReportsCmd(f))
	rootCmd.AddCommand(newOrganisationCmd(f))
	rootCmd.AddCommand(newSyncCmd(f))
	rootCmd.AddCommand(newConfigCmd(f))
	rootCmd.AddCommand(newRateLimitsCmd(f))

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		if silentErr, ok := err.(*cmdutil.SilentError); ok {
			return silentErr.Code
		}
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		return cmdutil.ExitError
	}

	return cmdutil.ExitOK
}
