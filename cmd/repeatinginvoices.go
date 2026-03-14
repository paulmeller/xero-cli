package cmd

import (
	"github.com/spf13/cobra"

	"github.com/paulmeller/xero-cli/internal/api"
	"github.com/paulmeller/xero-cli/internal/cmdutil"
	"github.com/paulmeller/xero-cli/internal/output"
)

func newRepeatingInvoicesCmd(f *cmdutil.Factory) *cobra.Command {
	def := cmdutil.ResourceDef{
		Name: "repeating-invoice", Plural: "repeating-invoices",
		APIPath: api.PathRepeatingInvoices, JSONKey: "RepeatingInvoices", IDField: "RepeatingInvoiceID",
		Columns: []output.Column{
			{Header: "ID", Path: "RepeatingInvoiceID"},
			{Header: "CONTACT", Path: "Contact.Name"},
			{Header: "TYPE", Path: "Type"},
			{Header: "STATUS", Path: "Status", Format: "status"},
			{Header: "TOTAL", Path: "Total", Format: "currency"},
			{Header: "SCHEDULE", Path: "Schedule.Period"},
		},
		ReadOnly:   true,
		HasHistory: true,
		HasAttach:  true,
	}
	return cmdutil.NewResourceCmd(f, def)
}
