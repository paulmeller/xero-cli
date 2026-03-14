package cmd

import (
	"github.com/spf13/cobra"

	"github.com/paulmeller/xero-cli/internal/api"
	"github.com/paulmeller/xero-cli/internal/cmdutil"
	"github.com/paulmeller/xero-cli/internal/output"
)

func newLinkedTransactionsCmd(f *cmdutil.Factory) *cobra.Command {
	def := cmdutil.ResourceDef{
		Name: "linked-transaction", Plural: "linked-transactions",
		APIPath: api.PathLinkedTransactions, JSONKey: "LinkedTransactions", IDField: "LinkedTransactionID",
		Columns: []output.Column{
			{Header: "ID", Path: "LinkedTransactionID"},
			{Header: "SOURCE", Path: "SourceTransactionID"},
			{Header: "TARGET", Path: "TargetTransactionID"},
			{Header: "CONTACT", Path: "ContactID"},
			{Header: "STATUS", Path: "Status", Format: "status"},
		},
		HasCreate: true, HasUpdate: true, HasDelete: true,
	}
	return cmdutil.NewResourceCmd(f, def)
}
