package cmd

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"

	"github.com/paulmeller/xero-cli/internal/api"
	"github.com/paulmeller/xero-cli/internal/cmdutil"
	"github.com/paulmeller/xero-cli/internal/output"
)

var invoiceColumns = []output.Column{
	{Header: "ID", Path: "InvoiceID"},
	{Header: "NUMBER", Path: "InvoiceNumber"},
	{Header: "CONTACT", Path: "Contact.Name"},
	{Header: "DATE", Path: "Date", Format: "date"},
	{Header: "DUE DATE", Path: "DueDate", Format: "date"},
	{Header: "STATUS", Path: "Status", Format: "status"},
	{Header: "TOTAL", Path: "Total", Format: "currency"},
}

func newInvoicesCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "invoices",
		Aliases: []string{"inv"},
		Short:   "Manage invoices",
	}

	cmd.AddCommand(newInvoicesListCmd(f))
	cmd.AddCommand(newInvoicesGetCmd(f))
	cmd.AddCommand(newInvoicesCreateCmd(f))
	cmd.AddCommand(newInvoicesUpdateCmd(f))
	cmd.AddCommand(newInvoicesDeleteCmd(f))
	cmd.AddCommand(newInvoicesVoidCmd(f))
	cmd.AddCommand(newInvoicesEmailCmd(f))
	cmd.AddCommand(newInvoicesOnlineURLCmd(f))
	cmd.AddCommand(newInvoicesHistoryCmd(f))
	cmd.AddCommand(newInvoicesAttachCmd(f))
	cmd.AddCommand(newInvoicesPDFCmd(f))

	return cmd
}

func newInvoicesListCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List invoices",
		Long: `List invoices with optional filtering and pagination.

Xero returns up to 100 records per page. Use --all to fetch all pages.`,
		Example: `  xero invoices list
  xero invoices list --all
  xero invoices list --status AUTHORISED,PAID
  xero invoices list --contact-id <id> --date-from 2025-01-01
  xero invoices list -o json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			format := cmdutil.GetOutputFormat(cmd, f.IO)
			allPages, _ := cmd.Flags().GetBool("all")

			// Try cache for --all with no filters
			if allPages {
				live, _ := cmd.Root().PersistentFlags().GetBool("live")
				hasFilters := cmdutil.HasChangedFilterFlags(cmd) || hasChangedInvoiceFilterFlags(cmd)
				if !live && !hasFilters {
					if data, ok := cmdutil.TryListCache(f, cmd, "Invoices", api.PathInvoices, "InvoiceID"); ok {
						items := gjson.ParseBytes(data).Get("Invoices")
						verbose, _ := cmd.Root().PersistentFlags().GetBool("verbose")
						if verbose {
							fmt.Fprintf(f.IO.ErrOut, "(cached) invoices\n")
						}
						formatter := f.Formatter(format)
						return formatter.FormatList(f.IO.Out, items, invoiceColumns)
					}
				}
			}

			client, err := f.APIClient()
			if err != nil {
				return err
			}
			cmdutil.ApplyClientFlags(cmd, client, f)

			params := cmdutil.BuildListParams(cmd)
			addInvoiceListParams(cmd, params)

			var items gjson.Result
			if allPages {
				pageSize, _ := cmd.Root().PersistentFlags().GetInt("page-size")
				items, err = api.PaginateAll(cmd.Context(), client, api.PathInvoices, params, "Invoices", pageSize)
				if err != nil {
					return err
				}
			} else {
				summary, _ := cmd.Flags().GetBool("summary")
				if summary {
					params.Set("summaryOnly", "true")
				}
				data, err := client.Get(cmd.Context(), api.PathInvoices, params)
				if err != nil {
					return err
				}
				items = gjson.ParseBytes(data).Get("Invoices")
			}

			formatter := f.Formatter(format)
			return formatter.FormatList(f.IO.Out, items, invoiceColumns)
		},
	}

	cmd.Flags().Bool("all", false, "Fetch all pages")
	cmd.Flags().StringSlice("status", nil, "Filter by status (DRAFT, SUBMITTED, AUTHORISED, PAID, VOIDED, DELETED)")
	cmd.Flags().String("contact-id", "", "Filter by contact ID")
	cmd.Flags().String("date-from", "", "Filter invoices from this date")
	cmd.Flags().String("date-to", "", "Filter invoices to this date")
	cmd.Flags().Bool("summary", false, "Return summary only (no line items)")
	cmd.Flags().StringSlice("numbers", nil, "Filter by invoice numbers")
	cmd.Flags().StringSlice("ids", nil, "Filter by invoice IDs")

	return cmd
}

func hasChangedInvoiceFilterFlags(cmd *cobra.Command) bool {
	for _, name := range []string{"status", "contact-id", "date-from", "date-to", "summary", "numbers", "ids"} {
		if cmd.Flags().Changed(name) {
			return true
		}
	}
	return false
}

func addInvoiceListParams(cmd *cobra.Command, params url.Values) {
	if statuses, _ := cmd.Flags().GetStringSlice("status"); len(statuses) > 0 {
		params.Set("Statuses", strings.Join(statuses, ","))
	}
	if v, _ := cmd.Flags().GetString("contact-id"); v != "" {
		params.Set("ContactIDs", v)
	}
	if numbers, _ := cmd.Flags().GetStringSlice("numbers"); len(numbers) > 0 {
		params.Set("InvoiceNumbers", strings.Join(numbers, ","))
	}
	if ids, _ := cmd.Flags().GetStringSlice("ids"); len(ids) > 0 {
		params.Set("IDs", strings.Join(ids, ","))
	}
	if v, _ := cmd.Flags().GetString("date-from"); v != "" {
		where := params.Get("where")
		clause := fmt.Sprintf("Date >= DateTime(%s)", xeroDateLiteral(v))
		if where != "" {
			where += " && " + clause
		} else {
			where = clause
		}
		params.Set("where", where)
	}
	if v, _ := cmd.Flags().GetString("date-to"); v != "" {
		where := params.Get("where")
		clause := fmt.Sprintf("Date <= DateTime(%s)", xeroDateLiteral(v))
		if where != "" {
			where += " && " + clause
		} else {
			where = clause
		}
		params.Set("where", where)
	}
}

func newInvoicesGetCmd(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "get <invoice-id>",
		Short: "Get an invoice by ID or number",
		Example: `  xero invoices get <invoice-id>
  xero invoices get INV-0001`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			format := cmdutil.GetOutputFormat(cmd, f.IO)

			// Try cache first
			live, _ := cmd.Root().PersistentFlags().GetBool("live")
			if !live {
				if data, ok := cmdutil.TryGetCache(f, cmd, "Invoices", api.PathInvoices, "InvoiceID", args[0]); ok {
					item := gjson.ParseBytes(data).Get("Invoices.0")
					if !item.Exists() {
						item = gjson.ParseBytes(data)
					}
					verbose, _ := cmd.Root().PersistentFlags().GetBool("verbose")
					if verbose {
						fmt.Fprintf(f.IO.ErrOut, "(cached) invoice %s\n", args[0])
					}
					formatter := f.Formatter(format)
					return formatter.FormatOne(f.IO.Out, item, invoiceColumns)
				}
			}

			client, err := f.APIClient()
			if err != nil {
				return err
			}
			cmdutil.ApplyClientFlags(cmd, client, f)

			path := fmt.Sprintf("%s/%s", api.PathInvoices, args[0])
			data, err := client.Get(cmd.Context(), path, nil)
			if err != nil {
				return err
			}

			formatter := f.Formatter(format)
			item := gjson.ParseBytes(data).Get("Invoices.0")
			if !item.Exists() {
				item = gjson.ParseBytes(data)
			}
			return formatter.FormatOne(f.IO.Out, item, invoiceColumns)
		},
	}
}

func newInvoicesCreateCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create an invoice",
		Long:  "Create an invoice from --file (JSON) or inline flags. Use --file - for stdin. JSON arrays create batch invoices (up to 50).",
		Example: `  xero invoices create --contact "Acme Corp" --line "Consulting,10,150,200"
  xero invoices create --file invoice.json
  cat invoice.json | xero invoices create --file -`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}
			cmdutil.ApplyClientFlags(cmd, client, f)

			idempotencyKey, _ := cmd.Flags().GetString("idempotency-key")

			// Validate mutual exclusivity of --file and inline flags
			fileFlag, _ := cmd.Flags().GetString("file")
			if fileFlag != "" {
				inlineFlags := []string{"type", "contact", "date", "due-date", "reference", "status", "currency", "line-amount-types", "line"}
				for _, name := range inlineFlags {
					if cmd.Flags().Changed(name) {
						return fmt.Errorf("--file cannot be combined with inline flags")
					}
				}
			}

			// Try file input first
			input, err := cmdutil.ReadInput(cmd)
			if err != nil {
				return err
			}

			if input != nil {
				// File/stdin input
				var result json.RawMessage
				if cmdutil.IsBatchInput(input) {
					wrapped := map[string]json.RawMessage{"Invoices": input}
					result, err = client.Post(cmd.Context(), api.PathInvoices, wrapped, idempotencyKey)
				} else {
					wrapped := map[string]json.RawMessage{"Invoices": json.RawMessage("[" + string(input) + "]")}
					result, err = client.Post(cmd.Context(), api.PathInvoices, wrapped, idempotencyKey)
				}
				if err != nil {
					return err
				}
				return outputInvoiceResult(f, cmd, result)
			}

			// Inline flag creation
			invoice, err := buildInvoiceFromFlags(cmd)
			if err != nil {
				return err
			}

			body := map[string][]api.Invoice{"Invoices": {*invoice}}
			result, err := client.Post(cmd.Context(), api.PathInvoices, body, idempotencyKey)
			if err != nil {
				return err
			}
			return outputInvoiceResult(f, cmd, result)
		},
	}

	cmd.Flags().String("file", "", "Input file path (use - for stdin)")
	cmd.Flags().String("idempotency-key", "", "Idempotency key for retry safety")
	cmd.Flags().String("type", "ACCREC", "Invoice type: ACCREC (sales) or ACCPAY (bills)")
	cmd.Flags().String("contact", "", "Contact name or ID")
	cmd.Flags().String("date", "", "Invoice date (YYYY-MM-DD)")
	cmd.Flags().String("due-date", "", "Due date (YYYY-MM-DD)")
	cmd.Flags().String("reference", "", "Reference")
	cmd.Flags().String("status", "", "Status (DRAFT, SUBMITTED, AUTHORISED)")
	cmd.Flags().String("currency", "", "Currency code")
	cmd.Flags().String("line-amount-types", "", "Line amount types: Exclusive, Inclusive, NoTax")
	cmd.Flags().StringArray("line", nil, `Line item: "Description,Qty,UnitAmount,AccountCode"`)

	return cmd
}

func newInvoicesUpdateCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <invoice-id>",
		Short: "Update an invoice",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}
			cmdutil.ApplyClientFlags(cmd, client, f)

			input, err := cmdutil.ReadInput(cmd)
			if err != nil {
				return err
			}

			if input != nil {
				// Wrap raw input in {"Invoices": [...]} envelope — Xero requires this
				wrapped := json.RawMessage(`{"Invoices":[` + string(input) + `]}`)

				path := fmt.Sprintf("%s/%s", api.PathInvoices, args[0])
				result, err := client.PostRaw(cmd.Context(), path, wrapped, "")
				if err != nil {
					return err
				}
				return outputInvoiceResult(f, cmd, result)
			}

			// Try inline flags
			status, _ := cmd.Flags().GetString("status")
			ref, _ := cmd.Flags().GetString("reference")
			dueDate, _ := cmd.Flags().GetString("due-date")

			if status == "" && ref == "" && dueDate == "" {
				return fmt.Errorf("provide --file or at least one of --status, --reference, --due-date")
			}

			body := map[string]any{}
			if status != "" {
				body["Status"] = status
			}
			if ref != "" {
				body["Reference"] = ref
			}
			if dueDate != "" {
				body["DueDate"] = dueDate
			}

			path := fmt.Sprintf("%s/%s", api.PathInvoices, args[0])
			result, err := client.Post(cmd.Context(), path, body, "")
			if err != nil {
				return err
			}
			return outputInvoiceResult(f, cmd, result)
		},
	}

	cmd.Flags().String("file", "", "Input file path (use - for stdin)")
	cmd.Flags().String("status", "", "Update status (DRAFT, SUBMITTED, AUTHORISED)")
	cmd.Flags().String("reference", "", "Update reference")
	cmd.Flags().String("due-date", "", "Update due date")

	return cmd
}

func newInvoicesDeleteCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <invoice-id>",
		Short: "Delete a draft invoice",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !cmdutil.ConfirmAction(f.IO, fmt.Sprintf("Delete invoice %s?", args[0]), cmd) {
				return fmt.Errorf("aborted; use --force to skip confirmation")
			}

			client, err := f.APIClient()
			if err != nil {
				return err
			}
			cmdutil.ApplyClientFlags(cmd, client, f)

			// Xero deletes invoices by setting Status to DELETED
			body := map[string]any{
				"InvoiceID": args[0],
				"Status":    "DELETED",
			}
			path := fmt.Sprintf("%s/%s", api.PathInvoices, args[0])
			_, err = client.Post(cmd.Context(), path, body, "")
			if err != nil {
				return err
			}

			quiet, _ := cmd.Root().PersistentFlags().GetBool("quiet")
			if !quiet {
				fmt.Fprintf(f.IO.ErrOut, "Deleted invoice %s\n", args[0])
			}
			return nil
		},
	}

	cmd.Flags().Bool("force", false, "Skip confirmation prompt")

	return cmd
}

func newInvoicesVoidCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "void <invoice-id>",
		Short: "Void an invoice",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !cmdutil.ConfirmAction(f.IO, fmt.Sprintf("Void invoice %s?", args[0]), cmd) {
				return fmt.Errorf("aborted; use --force to skip confirmation")
			}

			client, err := f.APIClient()
			if err != nil {
				return err
			}
			cmdutil.ApplyClientFlags(cmd, client, f)

			body := map[string]any{
				"InvoiceID": args[0],
				"Status":    "VOIDED",
			}
			path := fmt.Sprintf("%s/%s", api.PathInvoices, args[0])
			_, err = client.Post(cmd.Context(), path, body, "")
			if err != nil {
				return err
			}

			quiet, _ := cmd.Root().PersistentFlags().GetBool("quiet")
			if !quiet {
				fmt.Fprintf(f.IO.ErrOut, "Voided invoice %s\n", args[0])
			}
			return nil
		},
	}

	cmd.Flags().Bool("force", false, "Skip confirmation prompt")

	return cmd
}

func newInvoicesEmailCmd(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "email <invoice-id>",
		Short: "Email an invoice",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}
			cmdutil.ApplyClientFlags(cmd, client, f)

			path := fmt.Sprintf("%s/%s/Email", api.PathInvoices, args[0])
			_, err = client.Post(cmd.Context(), path, nil, "")
			if err != nil {
				return err
			}

			quiet, _ := cmd.Root().PersistentFlags().GetBool("quiet")
			if !quiet {
				fmt.Fprintf(f.IO.ErrOut, "Email sent for invoice %s\n", args[0])
			}
			return nil
		},
	}
}

func newInvoicesOnlineURLCmd(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "online-url <invoice-id>",
		Short: "Get the online invoice URL",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}
			cmdutil.ApplyClientFlags(cmd, client, f)

			path := fmt.Sprintf("%s/%s/OnlineInvoice", api.PathInvoices, args[0])
			data, err := client.Get(cmd.Context(), path, nil)
			if err != nil {
				return err
			}

			format := cmdutil.GetOutputFormat(cmd, f.IO)
			if format == "json" {
				formatter := f.Formatter("json")
				return formatter.FormatOne(f.IO.Out, gjson.ParseBytes(data), nil)
			}

			url := gjson.ParseBytes(data).Get("OnlineInvoices.0.OnlineInvoiceUrl").String()
			fmt.Fprintln(f.IO.Out, url)
			return nil
		},
	}
}

func newInvoicesHistoryCmd(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "history <invoice-id>",
		Short: "Get history for an invoice",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}
			cmdutil.ApplyClientFlags(cmd, client, f)

			path := fmt.Sprintf("%s/%s/History", api.PathInvoices, args[0])
			data, err := client.Get(cmd.Context(), path, nil)
			if err != nil {
				return err
			}

			format := cmdutil.GetOutputFormat(cmd, f.IO)
			formatter := f.Formatter(format)
			items := gjson.ParseBytes(data).Get("HistoryRecords")
			return formatter.FormatList(f.IO.Out, items, cmdutil.HistoryColumns)
		},
	}
}

func newInvoicesAttachCmd(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "attach <invoice-id> <file-path>",
		Short: "Attach a file to an invoice",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}
			cmdutil.ApplyClientFlags(cmd, client, f)

			filePath := args[1]
			fileName := filepath.Base(filePath)
			path := fmt.Sprintf("%s/%s/Attachments/%s", api.PathInvoices, args[0], fileName)

			data, err := os.ReadFile(filePath)
			if err != nil {
				return fmt.Errorf("cannot read file %s: %w", filePath, err)
			}

			contentType := cmdutil.DetectContentType(filePath)

			result, err := client.PutAttachment(cmd.Context(), path, data, contentType)
			if err != nil {
				return err
			}

			format := cmdutil.GetOutputFormat(cmd, f.IO)
			if format == "json" {
				formatter := f.Formatter("json")
				return formatter.FormatOne(f.IO.Out, gjson.ParseBytes(result), nil)
			}

			fmt.Fprintf(f.IO.ErrOut, "Attached %s to invoice %s\n", fileName, args[0])
			return nil
		},
	}
}

func buildInvoiceFromFlags(cmd *cobra.Command) (*api.Invoice, error) {
	inv := &api.Invoice{}

	inv.Type, _ = cmd.Flags().GetString("type")
	inv.Date, _ = cmd.Flags().GetString("date")
	inv.DueDate, _ = cmd.Flags().GetString("due-date")
	inv.Reference, _ = cmd.Flags().GetString("reference")
	inv.Status, _ = cmd.Flags().GetString("status")
	inv.CurrencyCode, _ = cmd.Flags().GetString("currency")
	inv.LineAmountTypes, _ = cmd.Flags().GetString("line-amount-types")

	contact, _ := cmd.Flags().GetString("contact")
	if contact != "" {
		inv.Contact = &api.Contact{Name: contact}
	}

	lines, _ := cmd.Flags().GetStringArray("line")
	for _, line := range lines {
		li, err := parseLineItem(line)
		if err != nil {
			return nil, err
		}
		inv.LineItems = append(inv.LineItems, li)
	}

	if inv.Contact == nil {
		return nil, fmt.Errorf("--contact is required")
	}

	return inv, nil
}

func parseLineItem(s string) (api.LineItem, error) {
	parts := strings.Split(s, ",")
	if len(parts) < 3 {
		return api.LineItem{}, fmt.Errorf("line item must have at least Description,Qty,UnitAmount: got %q", s)
	}

	li := api.LineItem{Description: strings.TrimSpace(parts[0])}

	var qty, amount float64
	if _, err := fmt.Sscanf(strings.TrimSpace(parts[1]), "%f", &qty); err != nil {
		return li, fmt.Errorf("invalid quantity %q in line item", parts[1])
	}
	li.Quantity = qty

	if _, err := fmt.Sscanf(strings.TrimSpace(parts[2]), "%f", &amount); err != nil {
		return li, fmt.Errorf("invalid unit amount %q in line item", parts[2])
	}
	li.UnitAmount = amount

	if len(parts) > 3 {
		li.AccountCode = strings.TrimSpace(parts[3])
	}

	return li, nil
}

func newInvoicesPDFCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pdf <invoice-id>",
		Short: "Download an invoice as PDF",
		Example: `  xero invoices pdf <invoice-id>
  xero invoices pdf <invoice-id> -o invoice.pdf`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}
			cmdutil.ApplyClientFlags(cmd, client, f)

			path := fmt.Sprintf("%s/%s", api.PathInvoices, args[0])
			data, err := client.GetPDF(cmd.Context(), path)
			if err != nil {
				return err
			}

			outPath, _ := cmd.Flags().GetString("out")
			if outPath == "" {
				// Default filename: INV-<id>.pdf
				outPath = fmt.Sprintf("invoice-%s.pdf", args[0])
			}

			if outPath == "-" {
				_, err = os.Stdout.Write(data)
				return err
			}

			if err := os.WriteFile(outPath, data, 0644); err != nil {
				return fmt.Errorf("cannot write PDF: %w", err)
			}

			quiet, _ := cmd.Root().PersistentFlags().GetBool("quiet")
			if !quiet {
				fmt.Fprintf(f.IO.ErrOut, "Saved %s (%d bytes)\n", outPath, len(data))
			}
			return nil
		},
	}

	cmd.Flags().String("out", "", "Output file path (default: invoice-<id>.pdf, use - for stdout)")

	return cmd
}

func outputInvoiceResult(f *cmdutil.Factory, cmd *cobra.Command, result json.RawMessage) error {
	return cmdutil.OutputResult(f, cmd, result, "Invoices", invoiceColumns)
}

// xeroDateLiteral converts "2026-01-15" to "2026,1,15" for Xero where clauses.
func xeroDateLiteral(date string) string {
	parts := strings.Split(date, "-")
	if len(parts) != 3 {
		return date
	}
	// Strip leading zeros
	y := strings.TrimLeft(parts[0], "0")
	m := strings.TrimLeft(parts[1], "0")
	d := strings.TrimLeft(parts[2], "0")
	return y + "," + m + "," + d
}
