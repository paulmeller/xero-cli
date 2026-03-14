package cmd

import (
	"github.com/spf13/cobra"

	"github.com/paulmeller/xero-cli/internal/api"
	"github.com/paulmeller/xero-cli/internal/cmdutil"
	"github.com/paulmeller/xero-cli/internal/output"
)

func newAccountsCmd(f *cmdutil.Factory) *cobra.Command {
	def := cmdutil.ResourceDef{
		Name:    "account",
		Plural:  "accounts",
		APIPath: api.PathAccounts,
		JSONKey: "Accounts",
		IDField: "AccountID",
		Columns: []output.Column{
			{Header: "ID", Path: "AccountID"},
			{Header: "CODE", Path: "Code"},
			{Header: "NAME", Path: "Name"},
			{Header: "TYPE", Path: "Type"},
			{Header: "CLASS", Path: "Class"},
			{Header: "STATUS", Path: "Status", Format: "status"},
			{Header: "TAX TYPE", Path: "TaxType"},
		},
		HasCreate:  true,
		HasUpdate:  true,
		HasDelete:  true,
		HasAttach:  true,
		HasArchive: true,
	}

	return cmdutil.NewResourceCmd(f, def)
}
