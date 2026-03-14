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

	cmd.AddCommand(newReportCmd(f, "profit-and-loss", "ProfitAndLoss"))
	cmd.AddCommand(newReportCmd(f, "balance-sheet", "BalanceSheet"))
	cmd.AddCommand(newReportCmd(f, "trial-balance", "TrialBalance"))
	cmd.AddCommand(newReportCmd(f, "aged-receivables", "AgedReceivablesByContact"))
	cmd.AddCommand(newReportCmd(f, "aged-payables", "AgedPayablesByContact"))
	cmd.AddCommand(newReportCmd(f, "bank-summary", "BankSummary"))
	cmd.AddCommand(newReportCmd(f, "budget-summary", "BudgetSummary"))
	cmd.AddCommand(newReportCmd(f, "executive-summary", "ExecutiveSummary"))
	cmd.AddCommand(newReportCmd(f, "gst", "GST"))
	cmd.AddCommand(newReportCmd(f, "1099", "TenNinetyNine"))

	return cmd
}

func newReportCmd(f *cmdutil.Factory, use string, reportID string) *cobra.Command {
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

			// Render report in table format
			return renderReport(f, data)
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

	return cmd
}

func renderReport(f *cmdutil.Factory, data []byte) error {
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
	title := report.Get("ReportName").String()
	if title != "" {
		fmt.Fprintf(f.IO.Out, "%s\n", title)
		fmt.Fprintf(f.IO.Out, "%s\n\n", strings.Repeat("=", len(title)))
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
			sectionTitle := row.Get("Title").String()
			if sectionTitle != "" {
				fmt.Fprintf(f.IO.Out, "\n%s\n", sectionTitle)
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

	formatter := f.Formatter("table")
	return formatter.FormatList(f.IO.Out, gjson.ParseBytes(rowsJSON), columns)
}
