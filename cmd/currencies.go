package cmd

import (
	"github.com/spf13/cobra"

	"github.com/paulmeller/xero-cli/internal/api"
	"github.com/paulmeller/xero-cli/internal/cmdutil"
	"github.com/paulmeller/xero-cli/internal/output"
)

func newCurrenciesCmd(f *cmdutil.Factory) *cobra.Command {
	def := cmdutil.ResourceDef{
		Name: "currency", Plural: "currencies",
		APIPath: api.PathCurrencies, JSONKey: "Currencies", IDField: "Code",
		Columns: []output.Column{
			{Header: "CODE", Path: "Code"},
			{Header: "DESCRIPTION", Path: "Description"},
		},
		HasCreate: true,
	}
	return cmdutil.NewResourceCmd(f, def)
}
