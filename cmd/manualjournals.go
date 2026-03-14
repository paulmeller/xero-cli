package cmd

import (
	"github.com/spf13/cobra"

	"github.com/paulmeller/xero-cli/internal/api"
	"github.com/paulmeller/xero-cli/internal/cmdutil"
	"github.com/paulmeller/xero-cli/internal/output"
)

func newManualJournalsCmd(f *cmdutil.Factory) *cobra.Command {
	def := cmdutil.ResourceDef{
		Name: "manual-journal", Plural: "manual-journals",
		APIPath: api.PathManualJournals, JSONKey: "ManualJournals", IDField: "ManualJournalID",
		Columns: []output.Column{
			{Header: "ID", Path: "ManualJournalID"},
			{Header: "NARRATION", Path: "Narration"},
			{Header: "DATE", Path: "Date", Format: "date"},
			{Header: "STATUS", Path: "Status", Format: "status"},
			{Header: "LINE COUNT", Path: "JournalLines.#"},
		},
		HasCreate: true, HasUpdate: true,
	}
	return cmdutil.NewResourceCmd(f, def)
}
