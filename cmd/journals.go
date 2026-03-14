package cmd

import (
	"github.com/spf13/cobra"

	"github.com/paulmeller/xero-cli/internal/api"
	"github.com/paulmeller/xero-cli/internal/cmdutil"
	"github.com/paulmeller/xero-cli/internal/output"
)

func newJournalsCmd(f *cmdutil.Factory) *cobra.Command {
	def := cmdutil.ResourceDef{
		Name: "journal", Plural: "journals",
		APIPath: api.PathJournals, JSONKey: "Journals", IDField: "JournalID",
		Columns: []output.Column{
			{Header: "ID", Path: "JournalID"},
			{Header: "NUMBER", Path: "JournalNumber"},
			{Header: "DATE", Path: "JournalDate", Format: "date"},
			{Header: "SOURCE", Path: "SourceType"},
			{Header: "REFERENCE", Path: "Reference"},
		},
		ReadOnly: true,
	}
	return cmdutil.NewResourceCmd(f, def)
}
