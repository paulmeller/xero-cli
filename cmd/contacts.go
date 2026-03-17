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

var contactColumns = []output.Column{
	{Header: "ID", Path: "ContactID"},
	{Header: "NAME", Path: "Name"},
	{Header: "EMAIL", Path: "EmailAddress"},
	{Header: "PHONE", Path: "Phones.0.PhoneNumber"},
	{Header: "STATUS", Path: "ContactStatus", Format: "status"},
	{Header: "BALANCE", Path: "Balances.AccountsReceivable.Outstanding", Format: "currency"},
}

func newContactsCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "contacts",
		Aliases: []string{"contact"},
		Short:   "Manage contacts",
	}

	cmd.AddCommand(newContactsListCmd(f))
	cmd.AddCommand(newContactsGetCmd(f))
	cmd.AddCommand(newContactsCreateCmd(f))
	cmd.AddCommand(newContactsUpdateCmd(f))
	cmd.AddCommand(newContactsArchiveCmd(f))
	cmd.AddCommand(newContactsHistoryCmd(f))
	cmd.AddCommand(newContactsAttachCmd(f))

	return cmd
}

func newContactsListCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List contacts",
		Long: `List contacts with optional filtering and pagination.

Xero returns up to 100 records per page. Use --all to fetch all pages.`,
		Example: `  xero contacts list
  xero contacts list --all
  xero contacts list --search "Acme"
  xero contacts list --is-customer -o json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			format := cmdutil.GetOutputFormat(cmd, f.IO)
			allPages, _ := cmd.Flags().GetBool("all")

			// Try cache for --all with no filters
			if allPages {
				live, _ := cmd.Root().PersistentFlags().GetBool("live")
				hasFilters := cmdutil.HasChangedFilterFlags(cmd) || hasChangedContactFilterFlags(cmd)
				if !live && !hasFilters {
					if data, ok := cmdutil.TryListCache(f, cmd, "Contacts", api.PathContacts, "ContactID"); ok {
						items := gjson.ParseBytes(data).Get("Contacts")
						verbose, _ := cmd.Root().PersistentFlags().GetBool("verbose")
						if verbose {
							fmt.Fprintf(f.IO.ErrOut, "(cached) contacts\n")
						}
						formatter := f.Formatter(format)
						return formatter.FormatList(f.IO.Out, items, contactColumns)
					}
				}
			}

			client, err := f.APIClient()
			if err != nil {
				return err
			}
			cmdutil.ApplyClientFlags(cmd, client, f)

			params := cmdutil.BuildListParams(cmd)
			addContactListParams(cmd, params)

			var items gjson.Result
			if allPages {
				pageSize, _ := cmd.Root().PersistentFlags().GetInt("page-size")
				items, err = api.PaginateAll(cmd.Context(), client, api.PathContacts, params, "Contacts", pageSize)
				if err != nil {
					return err
				}
			} else {
				summary, _ := cmd.Flags().GetBool("summary")
				if summary {
					params.Set("summaryOnly", "true")
				}
				data, err := client.Get(cmd.Context(), api.PathContacts, params)
				if err != nil {
					return err
				}
				items = gjson.ParseBytes(data).Get("Contacts")
			}

			formatter := f.Formatter(format)
			return formatter.FormatList(f.IO.Out, items, contactColumns)
		},
	}

	cmd.Flags().Bool("all", false, "Fetch all pages")
	cmd.Flags().String("search", "", "Search contacts by name, account number, etc.")
	cmd.Flags().StringSlice("ids", nil, "Filter by contact IDs")
	cmd.Flags().Bool("is-customer", false, "Filter to customers only")
	cmd.Flags().Bool("is-supplier", false, "Filter to suppliers only")
	cmd.Flags().Bool("include-archived", false, "Include archived contacts")
	cmd.Flags().Bool("summary", false, "Return summary only")

	return cmd
}

func hasChangedContactFilterFlags(cmd *cobra.Command) bool {
	for _, name := range []string{"search", "ids", "is-customer", "is-supplier", "include-archived", "summary"} {
		if cmd.Flags().Changed(name) {
			return true
		}
	}
	return false
}

func addContactListParams(cmd *cobra.Command, params url.Values) {
	if v, _ := cmd.Flags().GetString("search"); v != "" {
		params.Set("searchTerm", v)
	}
	if ids, _ := cmd.Flags().GetStringSlice("ids"); len(ids) > 0 {
		params.Set("IDs", strings.Join(ids, ","))
	}
	if v, _ := cmd.Flags().GetBool("is-customer"); v {
		params.Set("where", appendWhere(params.Get("where"), "IsCustomer==true"))
	}
	if v, _ := cmd.Flags().GetBool("is-supplier"); v {
		params.Set("where", appendWhere(params.Get("where"), "IsSupplier==true"))
	}
	if v, _ := cmd.Flags().GetBool("include-archived"); v {
		params.Set("includeArchived", "true")
	}
}

func newContactsGetCmd(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:     "get <contact-id>",
		Short:   "Get a contact by ID",
		Example: "  xero contacts get <contact-id>",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			format := cmdutil.GetOutputFormat(cmd, f.IO)

			// Try cache first
			live, _ := cmd.Root().PersistentFlags().GetBool("live")
			if !live {
				if data, ok := cmdutil.TryGetCache(f, cmd, "Contacts", api.PathContacts, "ContactID", args[0]); ok {
					item := gjson.ParseBytes(data).Get("Contacts.0")
					if !item.Exists() {
						item = gjson.ParseBytes(data)
					}
					verbose, _ := cmd.Root().PersistentFlags().GetBool("verbose")
					if verbose {
						fmt.Fprintf(f.IO.ErrOut, "(cached) contact %s\n", args[0])
					}
					formatter := f.Formatter(format)
					return formatter.FormatOne(f.IO.Out, item, contactColumns)
				}
			}

			client, err := f.APIClient()
			if err != nil {
				return err
			}
			cmdutil.ApplyClientFlags(cmd, client, f)

			path := fmt.Sprintf("%s/%s", api.PathContacts, args[0])
			data, err := client.Get(cmd.Context(), path, nil)
			if err != nil {
				return err
			}

			formatter := f.Formatter(format)
			item := gjson.ParseBytes(data).Get("Contacts.0")
			if !item.Exists() {
				item = gjson.ParseBytes(data)
			}
			return formatter.FormatOne(f.IO.Out, item, contactColumns)
		},
	}
}

func newContactsCreateCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a contact",
		Example: `  xero contacts create --name "Acme Corp" --email acme@example.com
  xero contacts create --file contact.json
  cat contact.json | xero contacts create --file -`,
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
				inlineFlags := []string{"name", "email", "phone", "tax-number"}
				for _, name := range inlineFlags {
					if cmd.Flags().Changed(name) {
						return fmt.Errorf("--file cannot be combined with inline flags")
					}
				}
			}

			input, err := cmdutil.ReadInput(cmd)
			if err != nil {
				return err
			}

			if input != nil {
				var result json.RawMessage
				if cmdutil.IsBatchInput(input) {
					wrapped := map[string]json.RawMessage{"Contacts": input}
					result, err = client.Post(cmd.Context(), api.PathContacts, wrapped, idempotencyKey)
				} else {
					wrapped := map[string]json.RawMessage{"Contacts": json.RawMessage("[" + string(input) + "]")}
					result, err = client.Post(cmd.Context(), api.PathContacts, wrapped, idempotencyKey)
				}
				if err != nil {
					return err
				}
				return outputContactResult(f, cmd, result)
			}

			// Inline flag creation
			contact := &api.Contact{}
			contact.Name, _ = cmd.Flags().GetString("name")
			contact.EmailAddress, _ = cmd.Flags().GetString("email")
			contact.TaxNumber, _ = cmd.Flags().GetString("tax-number")

			if phone, _ := cmd.Flags().GetString("phone"); phone != "" {
				contact.Phones = []api.Phone{{PhoneType: "DEFAULT", PhoneNumber: phone}}
			}

			if contact.Name == "" {
				return fmt.Errorf("--name is required")
			}

			body := map[string][]api.Contact{"Contacts": {*contact}}
			result, err := client.Post(cmd.Context(), api.PathContacts, body, idempotencyKey)
			if err != nil {
				return err
			}
			return outputContactResult(f, cmd, result)
		},
	}

	cmd.Flags().String("file", "", "Input file path (use - for stdin)")
	cmd.Flags().String("idempotency-key", "", "Idempotency key for retry safety")
	cmd.Flags().String("name", "", "Contact name")
	cmd.Flags().String("email", "", "Email address")
	cmd.Flags().String("phone", "", "Phone number")
	cmd.Flags().String("tax-number", "", "Tax number / ABN")

	return cmd
}

func newContactsUpdateCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <contact-id>",
		Short: "Update a contact",
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
			if input == nil {
				return fmt.Errorf("input required: use --file <path> or --file - for stdin")
			}

			// Wrap raw input in {"Contacts": [...]} envelope — Xero requires this
			wrapped := json.RawMessage(`{"Contacts":[` + string(input) + `]}`)

			path := fmt.Sprintf("%s/%s", api.PathContacts, args[0])
			result, err := client.PostRaw(cmd.Context(), path, wrapped, "")
			if err != nil {
				return err
			}
			return outputContactResult(f, cmd, result)
		},
	}

	cmd.Flags().String("file", "", "Input file path (use - for stdin)")

	return cmd
}

func newContactsArchiveCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "archive <contact-id>",
		Short: "Archive a contact",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !cmdutil.ConfirmAction(f.IO, fmt.Sprintf("Archive contact %s?", args[0]), cmd) {
				return fmt.Errorf("aborted; use --force to skip confirmation")
			}

			client, err := f.APIClient()
			if err != nil {
				return err
			}
			cmdutil.ApplyClientFlags(cmd, client, f)

			body := map[string]any{"ContactStatus": "ARCHIVED"}
			path := fmt.Sprintf("%s/%s", api.PathContacts, args[0])
			_, err = client.Post(cmd.Context(), path, body, "")
			if err != nil {
				return err
			}

			quiet, _ := cmd.Root().PersistentFlags().GetBool("quiet")
			if !quiet {
				fmt.Fprintf(f.IO.ErrOut, "Archived contact %s\n", args[0])
			}
			return nil
		},
	}

	cmd.Flags().Bool("force", false, "Skip confirmation prompt")

	return cmd
}

func newContactsHistoryCmd(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "history <contact-id>",
		Short: "Get history for a contact",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}
			cmdutil.ApplyClientFlags(cmd, client, f)

			path := fmt.Sprintf("%s/%s/History", api.PathContacts, args[0])
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

func newContactsAttachCmd(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "attach <contact-id> <file-path>",
		Short: "Attach a file to a contact",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}
			cmdutil.ApplyClientFlags(cmd, client, f)

			contactID := args[0]
			filePath := args[1]
			filename := filepath.Base(filePath)
			path := fmt.Sprintf("Contacts/%s/Attachments/%s", contactID, filename)

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

			fmt.Fprintf(f.IO.ErrOut, "Attached %s to contact %s\n", filename, contactID)
			return nil
		},
	}
}

func outputContactResult(f *cmdutil.Factory, cmd *cobra.Command, result json.RawMessage) error {
	return cmdutil.OutputResult(f, cmd, result, "Contacts", contactColumns)
}

func appendWhere(existing, clause string) string {
	if existing != "" {
		return existing + " && " + clause
	}
	return clause
}
