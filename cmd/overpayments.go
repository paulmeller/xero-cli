package cmd

import (
	"github.com/spf13/cobra"

	"github.com/paulmeller/xero-cli/internal/api"
	"github.com/paulmeller/xero-cli/internal/cmdutil"
	"github.com/paulmeller/xero-cli/internal/output"
)

func newOverpaymentsCmd(f *cmdutil.Factory) *cobra.Command {
	def := cmdutil.ResourceDef{
		Name: "overpayment", Plural: "overpayments",
		APIPath: api.PathOverpayments, JSONKey: "Overpayments", IDField: "OverpaymentID",
		Columns: []output.Column{
			{Header: "ID", Path: "OverpaymentID"},
			{Header: "CONTACT", Path: "Contact.Name"},
			{Header: "DATE", Path: "Date", Format: "date"},
			{Header: "STATUS", Path: "Status", Format: "status"},
			{Header: "TOTAL", Path: "Total", Format: "currency"},
			{Header: "REMAINING", Path: "RemainingCredit", Format: "currency"},
		},
		ReadOnly:    true,
		HasAllocate: true,
	}
	return cmdutil.NewResourceCmd(f, def)
}
