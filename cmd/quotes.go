package cmd

import (
	"github.com/spf13/cobra"

	"github.com/paulmeller/xero-cli/internal/api"
	"github.com/paulmeller/xero-cli/internal/cmdutil"
	"github.com/paulmeller/xero-cli/internal/output"
)

func newQuotesCmd(f *cmdutil.Factory) *cobra.Command {
	def := cmdutil.ResourceDef{
		Name: "quote", Plural: "quotes",
		APIPath: api.PathQuotes, JSONKey: "Quotes", IDField: "QuoteID",
		Columns: []output.Column{
			{Header: "ID", Path: "QuoteID"},
			{Header: "NUMBER", Path: "QuoteNumber"},
			{Header: "CONTACT", Path: "Contact.Name"},
			{Header: "DATE", Path: "Date", Format: "date"},
			{Header: "EXPIRY", Path: "ExpiryDate", Format: "date"},
			{Header: "STATUS", Path: "Status", Format: "status"},
			{Header: "TOTAL", Path: "Total", Format: "currency"},
		},
		HasCreate: true, HasUpdate: true, HasHistory: true,
	}
	return cmdutil.NewResourceCmd(f, def)
}
