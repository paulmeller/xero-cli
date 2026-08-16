package cmd

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"

	"github.com/paulmeller/xero-cli/internal/api"
	"github.com/paulmeller/xero-cli/internal/cmdutil"
	"github.com/paulmeller/xero-cli/internal/output"
)

func newReportsCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reports",
		Short: "Financial reports",
	}

	cmd.AddCommand(newReportCmd(f, "profit-and-loss", "ProfitAndLoss", false))
	cmd.AddCommand(newReportCmd(f, "balance-sheet", "BalanceSheet", false))
	cmd.AddCommand(newReportCmd(f, "trial-balance", "TrialBalance", false))
	cmd.AddCommand(newReportCmd(f, "aged-receivables", "AgedReceivablesByContact", false))
	cmd.AddCommand(newReportCmd(f, "aged-payables", "AgedPayablesByContact", false))
	cmd.AddCommand(newReportCmd(f, "bank-summary", "BankSummary", false))
	cmd.AddCommand(newReportCmd(f, "bank-statement", "BankStatement", true))
	cmd.AddCommand(newReportCmd(f, "budget-summary", "BudgetSummary", false))
	cmd.AddCommand(newReportCmd(f, "executive-summary", "ExecutiveSummary", false))
	cmd.AddCommand(newReportCmd(f, "gst", "GST", false))
	cmd.AddCommand(newReportCmd(f, "1099", "TenNinetyNine", false))

	return cmd
}

// needsBankAccount is true only for bank-statement, the one report Xero requires a
// bankAccountID for. Previously every report registered --bank-account-id unconditionally,
// so it showed up in --help for reports where it does nothing, and bank-statement never
// enforced that it was actually supplied.
func newReportCmd(f *cmdutil.Factory, use string, reportID string, needsBankAccount bool) *cobra.Command {
	cmd := &cobra.Command{
		Use:   use,
		Short: fmt.Sprintf("Run %s report", strings.ReplaceAll(use, "-", " ")),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}
			cmdutil.ApplyClientFlags(cmd, client, f)

			params := url.Values{}
			if v, _ := cmd.Flags().GetString("from-date"); v != "" {
				params.Set("fromDate", v)
			}
			if v, _ := cmd.Flags().GetString("to-date"); v != "" {
				params.Set("toDate", v)
			}
			if v, _ := cmd.Flags().GetString("periods"); v != "" {
				params.Set("periods", v)
			}
			if v, _ := cmd.Flags().GetString("timeframe"); v != "" {
				params.Set("timeframe", v)
			}
			if needsBankAccount {
				v, _ := cmd.Flags().GetString("bank-account-id")
				params.Set("bankAccountID", v)
			}
			if v, _ := cmd.Flags().GetString("tracking-category-id"); v != "" {
				params.Set("trackingCategoryID", v)
			}
			if v, _ := cmd.Flags().GetString("tracking-option-id"); v != "" {
				params.Set("trackingOptionID", v)
			}
			if v, _ := cmd.Flags().GetBool("standard-layout"); v {
				params.Set("standardLayout", "true")
			}
			if v, _ := cmd.Flags().GetBool("payments-only"); v {
				params.Set("paymentsOnly", "true")
			}

			path := fmt.Sprintf("%s/%s", api.PathReports, reportID)
			data, err := client.Get(cmd.Context(), path, params)
			if err != nil {
				return err
			}

			format := cmdutil.GetOutputFormat(cmd, f.IO)
			if format == "json" {
				formatter := f.Formatter("json")
				return formatter.FormatOne(f.IO.Out, gjson.ParseBytes(data), nil)
			}

			return renderReport(f, data, format)
		},
	}

	cmd.Flags().String("from-date", "", "Report start date (YYYY-MM-DD)")
	cmd.Flags().String("to-date", "", "Report end date (YYYY-MM-DD)")
	cmd.Flags().String("periods", "", "Number of periods")
	cmd.Flags().String("timeframe", "", "Period size: MONTH, QUARTER, YEAR")
	cmd.Flags().String("tracking-category-id", "", "Tracking category ID filter")
	cmd.Flags().String("tracking-option-id", "", "Tracking option ID filter")
	cmd.Flags().Bool("standard-layout", false, "Use standard layout")
	cmd.Flags().Bool("payments-only", false, "Cash basis")

	if needsBankAccount {
		cmd.Flags().String("bank-account-id", "", "Bank account ID (required)")
		cmd.MarkFlagRequired("bank-account-id")
	}

	return cmd
}

func renderReport(f *cmdutil.Factory, data []byte, format string) error {
	// Prose (title, section headers) is only safe to interleave with the data rows in the
	// human-readable table view - injecting free text into a csv/tsv stream would corrupt it.
	prose := format == "table" || format == ""

	parsed := gjson.ParseBytes(data)
	reports := parsed.Get("Reports")
	if !reports.Exists() {
		reports = parsed
	}

	var report gjson.Result
	if reports.IsArray() {
		arr := reports.Array()
		if len(arr) > 0 {
			report = arr[0]
		} else {
			fmt.Fprintln(f.IO.Out, "No report data.")
			return nil
		}
	} else {
		report = reports
	}

	// Print report title
	if prose {
		title := report.Get("ReportName").String()
		if title != "" {
			fmt.Fprintf(f.IO.Out, "%s\n", title)
			fmt.Fprintf(f.IO.Out, "%s\n\n", strings.Repeat("=", len(title)))
		}
	}

	// Build dynamic columns from the report's header row
	rows := report.Get("Rows")
	if !rows.Exists() || !rows.IsArray() {
		fmt.Fprintln(f.IO.Out, "No rows in report.")
		return nil
	}

	// Extract header columns from the first Header row
	var headers []string
	rows.ForEach(func(_, row gjson.Result) bool {
		rowType := row.Get("RowType").String()
		if rowType == "Header" {
			row.Get("Cells").ForEach(func(_, cell gjson.Result) bool {
				headers = append(headers, cell.Get("Value").String())
				return true
			})
			return false
		}
		return true
	})

	if len(headers) == 0 {
		headers = []string{"Account", "Value"}
	}

	// Build columns
	columns := make([]output.Column, len(headers))
	for i, h := range headers {
		columns[i] = output.Column{
			Header: strings.ToUpper(h),
			Path:   fmt.Sprintf("Cells.%d.Value", i),
		}
	}

	// Collect data rows
	var dataRows []gjson.Result
	rows.ForEach(func(_, row gjson.Result) bool {
		rowType := row.Get("RowType").String()
		switch rowType {
		case "Section":
			if prose {
				sectionTitle := row.Get("Title").String()
				if sectionTitle != "" {
					fmt.Fprintf(f.IO.Out, "\n%s\n", sectionTitle)
				}
			}
			row.Get("Rows").ForEach(func(_, subRow gjson.Result) bool {
				dataRows = append(dataRows, subRow)
				return true
			})
		case "Row":
			dataRows = append(dataRows, row)
		case "SummaryRow":
			dataRows = append(dataRows, row)
		}
		return true
	})

	// Build a JSON array of the rows for the formatter
	var rowsJSON []byte
	rowsJSON = append(rowsJSON, '[')
	for i, r := range dataRows {
		if i > 0 {
			rowsJSON = append(rowsJSON, ',')
		}
		rowsJSON = append(rowsJSON, []byte(r.Raw)...)
	}
	rowsJSON = append(rowsJSON, ']')

	formatter := f.Formatter(format)
	return formatter.FormatList(f.IO.Out, gjson.ParseBytes(rowsJSON), columns)
}
