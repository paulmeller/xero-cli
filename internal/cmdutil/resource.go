package cmdutil

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
	"github.com/paulmeller/xero-cli/internal/output"
)

// ResourceDef defines a Xero resource for auto-generated CRUD commands.
type ResourceDef struct {
	Name          string          // singular, e.g. "invoice"
	Plural        string          // e.g. "invoices"
	APIPath       string          // e.g. "Invoices"
	JSONKey       string          // response wrapper key, e.g. "Invoices"
	IDField       string          // e.g. "InvoiceID"
	Columns       []output.Column // columns for list/table display
	DetailColumns []output.Column // columns for single-item display (if different)
	ListFlags     func(cmd *cobra.Command)
	ListParams    func(cmd *cobra.Command, params url.Values) // extra list params from flags
	HasCreate     bool
	HasUpdate     bool
	HasDelete     bool
	HasHistory    bool
	HasAttach     bool
	HasAllocate   bool
	HasArchive    bool
	ReadOnly      bool
}

// ListOpts configures a list command.
type ListOpts struct {
	Factory  *Factory
	Def      ResourceDef
	AllPages bool
}

// NewResourceCmd creates a parent command with all sub-commands for a resource.
func NewResourceCmd(f *Factory, def ResourceDef) *cobra.Command {
	cmd := &cobra.Command{
		Use:   def.Plural,
		Short: fmt.Sprintf("Manage %s", def.Plural),
	}

	cmd.AddCommand(NewListCmd(f, def))
	cmd.AddCommand(NewGetCmd(f, def))

	if def.HasCreate && !def.ReadOnly {
		cmd.AddCommand(NewCreateCmd(f, def))
	}
	if def.HasUpdate && !def.ReadOnly {
		cmd.AddCommand(NewUpdateCmd(f, def))
	}
	if def.HasDelete && !def.ReadOnly {
		cmd.AddCommand(NewDeleteCmd(f, def))
	}
	if def.HasHistory {
		cmd.AddCommand(NewHistoryCmd(f, def))
	}
	if def.HasAllocate {
		cmd.AddCommand(NewAllocateCmd(f, def))
	}
	if def.HasAttach {
		cmd.AddCommand(NewAttachCmd(f, def))
	}
	if def.HasArchive {
		cmd.AddCommand(NewArchiveCmd(f, def))
	}

	return cmd
}

// NewListCmd creates a list sub-command for a resource.
func NewListCmd(f *Factory, def ResourceDef) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: fmt.Sprintf("List %s", def.Plural),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}
			ApplyClientFlags(cmd, client, f)

			params := BuildListParams(cmd)
			if def.ListParams != nil {
				def.ListParams(cmd, params)
			}

			format := GetOutputFormat(cmd, f.IO)

			allPages, _ := cmd.Flags().GetBool("all")
			var items gjson.Result
			if allPages {
				pageSize, _ := cmd.Root().PersistentFlags().GetInt("page-size")
				items, err = api.PaginateAll(cmd.Context(), client, def.APIPath, params, def.JSONKey, pageSize)
				if err != nil {
					return err
				}
			} else {
				data, err := client.Get(cmd.Context(), def.APIPath, params)
				if err != nil {
					return err
				}
				items = gjson.ParseBytes(data).Get(def.JSONKey)
			}

			formatter := f.Formatter(format)
			return formatter.FormatList(f.IO.Out, items, def.Columns)
		},
	}

	cmd.Flags().Bool("all", false, "Fetch all pages")
	if def.ListFlags != nil {
		def.ListFlags(cmd)
	}

	return cmd
}

// NewGetCmd creates a get sub-command for a resource.
func NewGetCmd(f *Factory, def ResourceDef) *cobra.Command {
	return &cobra.Command{
		Use:   fmt.Sprintf("get <%s-id>", def.Name),
		Short: fmt.Sprintf("Get a %s by ID", def.Name),
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}
			ApplyClientFlags(cmd, client, f)

			path := fmt.Sprintf("%s/%s", def.APIPath, args[0])
			data, err := client.Get(cmd.Context(), path, nil)
			if err != nil {
				return err
			}

			format := GetOutputFormat(cmd, f.IO)
			parsed := gjson.ParseBytes(data)

			// Try to get the single item from the wrapper
			item := parsed.Get(def.JSONKey + ".0")
			if !item.Exists() {
				item = parsed
			}

			cols := def.DetailColumns
			if cols == nil {
				cols = def.Columns
			}

			formatter := f.Formatter(format)
			return formatter.FormatOne(f.IO.Out, item, cols)
		},
	}
}

// NewCreateCmd creates a create sub-command for a resource.
func NewCreateCmd(f *Factory, def ResourceDef) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: fmt.Sprintf("Create a %s", def.Name),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}
			ApplyClientFlags(cmd, client, f)

			input, err := ReadInput(cmd)
			if err != nil {
				return err
			}
			if input == nil {
				return fmt.Errorf("input required: use --file <path> or --file - for stdin")
			}

			idempotencyKey, _ := cmd.Flags().GetString("idempotency-key")

			var result json.RawMessage
			if IsBatchInput(input) {
				// Wrap array in the resource key for Xero API
				wrapped := map[string]json.RawMessage{def.JSONKey: input}
				result, err = client.Post(cmd.Context(), def.APIPath, wrapped, idempotencyKey)
			} else {
				// Single item - wrap in the resource key
				wrapped := map[string]json.RawMessage{def.JSONKey: json.RawMessage("[" + string(input) + "]")}
				result, err = client.Post(cmd.Context(), def.APIPath, wrapped, idempotencyKey)
			}
			if err != nil {
				return err
			}

			format := GetOutputFormat(cmd, f.IO)
			formatter := f.Formatter(format)

			parsed := gjson.ParseBytes(result)
			items := parsed.Get(def.JSONKey)
			if items.IsArray() && len(items.Array()) == 1 {
				return formatter.FormatOne(f.IO.Out, items.Array()[0], def.Columns)
			}
			return formatter.FormatList(f.IO.Out, items, def.Columns)
		},
	}

	cmd.Flags().String("file", "", "Input file path (use - for stdin)")
	cmd.Flags().String("idempotency-key", "", "Idempotency key for retry safety")

	return cmd
}

// NewUpdateCmd creates an update sub-command for a resource.
func NewUpdateCmd(f *Factory, def ResourceDef) *cobra.Command {
	cmd := &cobra.Command{
		Use:   fmt.Sprintf("update <%s-id>", def.Name),
		Short: fmt.Sprintf("Update a %s", def.Name),
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}
			ApplyClientFlags(cmd, client, f)

			input, err := ReadInput(cmd)
			if err != nil {
				return err
			}
			if input == nil {
				return fmt.Errorf("input required: use --file <path> or --file - for stdin")
			}

			// Wrap raw input in the resource envelope (e.g. {"Invoices": [...]})
			// Xero requires this wrapper for updates.
			wrapped := json.RawMessage(`{"` + def.JSONKey + `":[` + string(input) + `]}`)

			path := fmt.Sprintf("%s/%s", def.APIPath, args[0])
			result, err := client.PostRaw(cmd.Context(), path, wrapped, "")
			if err != nil {
				return err
			}

			format := GetOutputFormat(cmd, f.IO)
			formatter := f.Formatter(format)

			parsed := gjson.ParseBytes(result)
			item := parsed.Get(def.JSONKey + ".0")
			if !item.Exists() {
				item = parsed
			}
			return formatter.FormatOne(f.IO.Out, item, def.Columns)
		},
	}

	cmd.Flags().String("file", "", "Input file path (use - for stdin)")

	return cmd
}

// NewDeleteCmd creates a delete sub-command for a resource.
func NewDeleteCmd(f *Factory, def ResourceDef) *cobra.Command {
	cmd := &cobra.Command{
		Use:   fmt.Sprintf("delete <%s-id>", def.Name),
		Short: fmt.Sprintf("Delete a %s", def.Name),
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !ConfirmAction(f.IO, fmt.Sprintf("Delete %s %s?", def.Name, args[0]), cmd) {
				return fmt.Errorf("aborted; use --force to skip confirmation")
			}

			client, err := f.APIClient()
			if err != nil {
				return err
			}
			ApplyClientFlags(cmd, client, f)

			path := fmt.Sprintf("%s/%s", def.APIPath, args[0])
			_, err = client.Delete(cmd.Context(), path)
			if err != nil {
				return err
			}

			quiet, _ := cmd.Root().PersistentFlags().GetBool("quiet")
			if !quiet {
				fmt.Fprintf(f.IO.ErrOut, "Deleted %s %s\n", def.Name, args[0])
			}
			return nil
		},
	}

	cmd.Flags().Bool("force", false, "Skip confirmation prompt")

	return cmd
}

// NewHistoryCmd creates a history sub-command for a resource.
func NewHistoryCmd(f *Factory, def ResourceDef) *cobra.Command {
	return &cobra.Command{
		Use:   fmt.Sprintf("history <%s-id>", def.Name),
		Short: fmt.Sprintf("Get history for a %s", def.Name),
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}
			ApplyClientFlags(cmd, client, f)

			path := fmt.Sprintf("%s/%s/History", def.APIPath, args[0])
			data, err := client.Get(cmd.Context(), path, nil)
			if err != nil {
				return err
			}

			format := GetOutputFormat(cmd, f.IO)
			formatter := f.Formatter(format)

			items := gjson.ParseBytes(data).Get("HistoryRecords")

			histCols := []output.Column{
				{Header: "DATE", Path: "DateUTCString", Format: "date"},
				{Header: "USER", Path: "User"},
				{Header: "CHANGES", Path: "Changes"},
				{Header: "DETAILS", Path: "Details"},
			}

			return formatter.FormatList(f.IO.Out, items, histCols)
		},
	}
}

// NewAllocateCmd creates an allocate sub-command (for credit notes, overpayments, prepayments).
func NewAllocateCmd(f *Factory, def ResourceDef) *cobra.Command {
	cmd := &cobra.Command{
		Use:   fmt.Sprintf("allocate <%s-id>", def.Name),
		Short: fmt.Sprintf("Allocate a %s to an invoice", def.Name),
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}
			ApplyClientFlags(cmd, client, f)

			invoiceID, _ := cmd.Flags().GetString("invoice")
			amount, _ := cmd.Flags().GetFloat64("amount")

			if invoiceID == "" || amount == 0 {
				return fmt.Errorf("--invoice and --amount are required")
			}

			body := map[string]any{
				"Allocations": []map[string]any{
					{
						"Invoice": map[string]string{"InvoiceID": invoiceID},
						"Amount":  fmt.Sprintf("%.2f", amount),
					},
				},
			}

			path := fmt.Sprintf("%s/%s/Allocations", def.APIPath, args[0])
			result, err := client.Put(cmd.Context(), path, body)
			if err != nil {
				return err
			}

			format := GetOutputFormat(cmd, f.IO)
			formatter := f.Formatter(format)
			return formatter.FormatOne(f.IO.Out, gjson.ParseBytes(result), nil)
		},
	}

	cmd.Flags().String("invoice", "", "Invoice ID to allocate to")
	cmd.Flags().Float64("amount", 0, "Amount to allocate")

	return cmd
}

// NewAttachCmd creates an attach sub-command for a resource.
func NewAttachCmd(f *Factory, def ResourceDef) *cobra.Command {
	return &cobra.Command{
		Use:   fmt.Sprintf("attach <%s-id> <file-path>", def.Name),
		Short: fmt.Sprintf("Attach a file to a %s", def.Name),
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}
			ApplyClientFlags(cmd, client, f)

			filePath := args[1]
			path := fmt.Sprintf("%s/%s/Attachments/%s", def.APIPath, args[0], filepath.Base(filePath))

			data, err := os.ReadFile(filePath)
			if err != nil {
				return fmt.Errorf("cannot read file %s: %w", filePath, err)
			}

			contentType := detectContentType(filePath)
			result, err := client.PutAttachment(cmd.Context(), path, data, contentType)
			if err != nil {
				return err
			}

			format := GetOutputFormat(cmd, f.IO)
			formatter := f.Formatter(format)
			return formatter.FormatOne(f.IO.Out, gjson.ParseBytes(result), nil)
		},
	}
}

// NewArchiveCmd creates an archive sub-command for a resource.
func NewArchiveCmd(f *Factory, def ResourceDef) *cobra.Command {
	cmd := &cobra.Command{
		Use:   fmt.Sprintf("archive <%s-id>", def.Name),
		Short: fmt.Sprintf("Archive a %s", def.Name),
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !ConfirmAction(f.IO, fmt.Sprintf("Archive %s %s?", def.Name, args[0]), cmd) {
				return fmt.Errorf("aborted; use --force to skip confirmation")
			}

			client, err := f.APIClient()
			if err != nil {
				return err
			}
			ApplyClientFlags(cmd, client, f)

			body := map[string]any{
				def.IDField: args[0],
				"Status":    "ARCHIVED",
			}

			path := fmt.Sprintf("%s/%s", def.APIPath, args[0])
			result, err := client.Put(cmd.Context(), path, body)
			if err != nil {
				return err
			}

			format := GetOutputFormat(cmd, f.IO)
			formatter := f.Formatter(format)
			return formatter.FormatOne(f.IO.Out, gjson.ParseBytes(result), def.Columns)
		},
	}

	cmd.Flags().Bool("force", false, "Skip confirmation prompt")

	return cmd
}

func detectContentType(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".pdf":
		return "application/pdf"
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".gif":
		return "image/gif"
	case ".csv":
		return "text/csv"
	case ".xlsx":
		return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	case ".doc":
		return "application/msword"
	case ".docx":
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	default:
		return "application/octet-stream"
	}
}

// ApplyClientFlags reads global flags and configures the API client.
func ApplyClientFlags(cmd *cobra.Command, client *api.Client, f *Factory) {
	verbose, _ := cmd.Root().PersistentFlags().GetBool("verbose")
	if verbose {
		client.SetVerbose(true, f.IO.ErrOut)
	}

	dryRun, _ := cmd.Root().PersistentFlags().GetBool("dry-run")
	if dryRun {
		client.SetDryRun(true)
	}

	tenant, _ := cmd.Root().PersistentFlags().GetString("tenant")
	if tenant != "" {
		client.SetTenantID(tenant)
	}
}
