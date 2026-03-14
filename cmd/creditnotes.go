package cmd

import (
	"github.com/spf13/cobra"

	"github.com/paulmeller/xero-cli/internal/api"
	"github.com/paulmeller/xero-cli/internal/cmdutil"
	"github.com/paulmeller/xero-cli/internal/output"
)

func newCreditNotesCmd(f *cmdutil.Factory) *cobra.Command {
	def := cmdutil.ResourceDef{
		Name:    "credit-note",
		Plural:  "credit-notes",
		APIPath: api.PathCreditNotes,
		JSONKey: "CreditNotes",
		IDField: "CreditNoteID",
		Columns: []output.Column{
			{Header: "ID", Path: "CreditNoteID"},
			{Header: "NUMBER", Path: "CreditNoteNumber"},
			{Header: "CONTACT", Path: "Contact.Name"},
			{Header: "DATE", Path: "Date", Format: "date"},
			{Header: "STATUS", Path: "Status", Format: "status"},
			{Header: "TOTAL", Path: "Total", Format: "currency"},
			{Header: "REMAINING", Path: "RemainingCredit", Format: "currency"},
		},
		HasCreate:   true,
		HasUpdate:   true,
		HasAllocate: true,
		HasAttach:   true,
	}

	return cmdutil.NewResourceCmd(f, def)
}
