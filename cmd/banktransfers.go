package cmd

import (
	"github.com/spf13/cobra"

	"github.com/paulmeller/xero-cli/internal/api"
	"github.com/paulmeller/xero-cli/internal/cmdutil"
	"github.com/paulmeller/xero-cli/internal/output"
)

func newBankTransfersCmd(f *cmdutil.Factory) *cobra.Command {
	def := cmdutil.ResourceDef{
		Name:    "bank-transfer",
		Plural:  "bank-transfers",
		APIPath: api.PathBankTransfers,
		JSONKey: "BankTransfers",
		IDField: "BankTransferID",
		Columns: []output.Column{
			{Header: "ID", Path: "BankTransferID"},
			{Header: "DATE", Path: "Date", Format: "date"},
			{Header: "AMOUNT", Path: "Amount", Format: "currency"},
			{Header: "FROM", Path: "FromBankAccount.Name"},
			{Header: "TO", Path: "ToBankAccount.Name"},
		},
		HasCreate:     true,
		HasHistory:    true,
		HasAttach:     true,
		CreateUsesPut: true,
	}

	return cmdutil.NewResourceCmd(f, def)
}
