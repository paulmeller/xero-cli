package cmd

import (
	"github.com/spf13/cobra"

	"github.com/paulmeller/xero-cli/internal/api"
	"github.com/paulmeller/xero-cli/internal/cmdutil"
	"github.com/paulmeller/xero-cli/internal/output"
)

func newBankTransactionsCmd(f *cmdutil.Factory) *cobra.Command {
	def := cmdutil.ResourceDef{
		Name:    "bank-transaction",
		Plural:  "bank-transactions",
		APIPath: api.PathBankTransactions,
		JSONKey: "BankTransactions",
		IDField: "BankTransactionID",
		Columns: []output.Column{
			{Header: "ID", Path: "BankTransactionID"},
			{Header: "TYPE", Path: "Type"},
			{Header: "CONTACT", Path: "Contact.Name"},
			{Header: "DATE", Path: "Date", Format: "date"},
			{Header: "STATUS", Path: "Status", Format: "status"},
			{Header: "TOTAL", Path: "Total", Format: "currency"},
			{Header: "ACCOUNT", Path: "BankAccount.Name"},
		},
		HasCreate:  true,
		HasUpdate:  true,
		HasHistory: true,
		HasAttach:  true,
	}

	return cmdutil.NewResourceCmd(f, def)
}
