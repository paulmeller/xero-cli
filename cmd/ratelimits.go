package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/paulmeller/xero-cli/internal/cmdutil"
)

func newRateLimitsCmd(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:     "rate-limits",
		Aliases: []string{"limits"},
		Short:   "Show current API rate limit status",
		Long:    "Makes a lightweight API call and reports the rate limit headers from the response.",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}
			cmdutil.ApplyClientFlags(cmd, client, f)

			// Use a lightweight call — GET /Organisation
			data, headers, err := client.GetWithHeaders(cmd.Context(), "Organisation", nil)
			if err != nil {
				return err
			}
			_ = data

			format := cmdutil.GetOutputFormat(cmd, f.IO)
			if format == "json" {
				fmt.Fprintf(f.IO.Out, `{"minute_limit_remaining":%q,"daily_limit_remaining":%q,"app_minute_limit_remaining":%q}`+"\n",
					headers.Get("X-MinLimit-Remaining"),
					headers.Get("X-DayLimit-Remaining"),
					headers.Get("X-AppMinLimit-Remaining"),
				)
				return nil
			}

			fmt.Fprintf(f.IO.Out, "Minute limit remaining:     %s\n", valueOrNA(headers.Get("X-MinLimit-Remaining")))
			fmt.Fprintf(f.IO.Out, "Daily limit remaining:      %s\n", valueOrNA(headers.Get("X-DayLimit-Remaining")))
			fmt.Fprintf(f.IO.Out, "App minute limit remaining: %s\n", valueOrNA(headers.Get("X-AppMinLimit-Remaining")))
			return nil
		},
	}
}

func valueOrNA(s string) string {
	if s == "" {
		return "N/A"
	}
	return s
}
