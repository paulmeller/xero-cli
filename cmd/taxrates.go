package cmd

import (
	"github.com/spf13/cobra"

	"github.com/paulmeller/xero-cli/internal/api"
	"github.com/paulmeller/xero-cli/internal/cmdutil"
	"github.com/paulmeller/xero-cli/internal/output"
)

func newTaxRatesCmd(f *cmdutil.Factory) *cobra.Command {
	def := cmdutil.ResourceDef{
		Name: "tax-rate", Plural: "tax-rates",
		APIPath: api.PathTaxRates, JSONKey: "TaxRates", IDField: "TaxType",
		Columns: []output.Column{
			{Header: "NAME", Path: "Name"},
			{Header: "TAX TYPE", Path: "TaxType"},
			{Header: "RATE", Path: "EffectiveRate"},
			{Header: "REPORT TYPE", Path: "ReportTaxType"},
			{Header: "STATUS", Path: "Status", Format: "status"},
		},
		HasCreate: true, HasUpdate: true,
	}
	return cmdutil.NewResourceCmd(f, def)
}
