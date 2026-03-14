package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"

	"github.com/paulmeller/xero-cli/internal/api"
	"github.com/paulmeller/xero-cli/internal/cmdutil"
	"github.com/paulmeller/xero-cli/internal/output"
)

var paymentColumns = []output.Column{
	{Header: "ID", Path: "PaymentID"},
	{Header: "DATE", Path: "Date", Format: "date"},
	{Header: "AMOUNT", Path: "Amount", Format: "currency"},
	{Header: "REFERENCE", Path: "Reference"},
	{Header: "INVOICE", Path: "Invoice.InvoiceNumber"},
	{Header: "ACCOUNT", Path: "Account.Code"},
	{Header: "STATUS", Path: "Status", Format: "status"},
}

func newPaymentsCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "payments",
		Aliases: []string{"payment"},
		Short:   "Manage payments",
	}

	def := cmdutil.ResourceDef{
		Name:      "payment",
		Plural:    "payments",
		APIPath:   api.PathPayments,
		JSONKey:   "Payments",
		IDField:   "PaymentID",
		Columns:   paymentColumns,
		HasCreate: true,
		HasDelete: true,
	}

	cmd.AddCommand(cmdutil.NewListCmd(f, def))
	cmd.AddCommand(cmdutil.NewGetCmd(f, def))
	cmd.AddCommand(newPaymentCreateCmd(f))
	cmd.AddCommand(newPaymentDeleteCmd(f))

	return cmd
}

func newPaymentCreateCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a payment",
		Example: `  xero payments create --invoice INV-0001 --account 090 --amount 500.00
  xero payments create --invoice INV-0001 --account 090 --amount 500.00 --date 2025-03-01
  xero payments create --file payment.json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}
			cmdutil.ApplyClientFlags(cmd, client, f)

			idempotencyKey, _ := cmd.Flags().GetString("idempotency-key")

			input, err := cmdutil.ReadInput(cmd)
			if err != nil {
				return err
			}

			if input != nil {
				wrapped := map[string]json.RawMessage{"Payments": json.RawMessage("[" + string(input) + "]")}
				if cmdutil.IsBatchInput(input) {
					wrapped = map[string]json.RawMessage{"Payments": input}
				}
				result, err := client.Post(cmd.Context(), api.PathPayments, wrapped, idempotencyKey)
				if err != nil {
					return err
				}
				format := cmdutil.GetOutputFormat(cmd, f.IO)
				formatter := f.Formatter(format)
				return formatter.FormatOne(f.IO.Out, gjson.ParseBytes(result).Get("Payments.0"), paymentColumns)
			}

			// Inline creation
			invoiceNum, _ := cmd.Flags().GetString("invoice")
			accountCode, _ := cmd.Flags().GetString("account")
			amount, _ := cmd.Flags().GetFloat64("amount")
			date, _ := cmd.Flags().GetString("date")
			ref, _ := cmd.Flags().GetString("reference")

			if invoiceNum == "" || accountCode == "" || amount == 0 {
				return fmt.Errorf("--invoice, --account, and --amount are required")
			}

			payment := api.Payment{
				Invoice:   &api.Invoice{InvoiceNumber: invoiceNum},
				Account:   &api.Account{Code: accountCode},
				Amount:    amount,
				Date:      date,
				Reference: ref,
			}

			body := map[string][]api.Payment{"Payments": {payment}}
			result, err := client.Post(cmd.Context(), api.PathPayments, body, idempotencyKey)
			if err != nil {
				return err
			}

			format := cmdutil.GetOutputFormat(cmd, f.IO)
			formatter := f.Formatter(format)
			return formatter.FormatOne(f.IO.Out, gjson.ParseBytes(result).Get("Payments.0"), paymentColumns)
		},
	}

	cmd.Flags().String("file", "", "Input file path (use - for stdin)")
	cmd.Flags().String("idempotency-key", "", "Idempotency key")
	cmd.Flags().String("invoice", "", "Invoice number")
	cmd.Flags().String("account", "", "Account code")
	cmd.Flags().Float64("amount", 0, "Payment amount")
	cmd.Flags().String("date", "", "Payment date (YYYY-MM-DD)")
	cmd.Flags().String("reference", "", "Payment reference")

	return cmd
}

func newPaymentDeleteCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <payment-id>",
		Short: "Delete a payment",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !cmdutil.ConfirmAction(f.IO, fmt.Sprintf("Delete payment %s?", args[0]), cmd) {
				return fmt.Errorf("aborted; use --force to skip confirmation")
			}

			client, err := f.APIClient()
			if err != nil {
				return err
			}
			cmdutil.ApplyClientFlags(cmd, client, f)

			body := map[string]string{"Status": "DELETED"}
			path := fmt.Sprintf("%s/%s", api.PathPayments, args[0])
			_, err = client.Post(cmd.Context(), path, body, "")
			if err != nil {
				return err
			}

			quiet, _ := cmd.Root().PersistentFlags().GetBool("quiet")
			if !quiet {
				fmt.Fprintf(f.IO.ErrOut, "Deleted payment %s\n", args[0])
			}
			return nil
		},
	}

	cmd.Flags().Bool("force", false, "Skip confirmation prompt")
	return cmd
}
