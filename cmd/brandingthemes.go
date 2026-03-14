package cmd

import (
	"github.com/spf13/cobra"

	"github.com/paulmeller/xero-cli/internal/api"
	"github.com/paulmeller/xero-cli/internal/cmdutil"
	"github.com/paulmeller/xero-cli/internal/output"
)

func newBrandingThemesCmd(f *cmdutil.Factory) *cobra.Command {
	def := cmdutil.ResourceDef{
		Name: "branding-theme", Plural: "branding-themes",
		APIPath: api.PathBrandingThemes, JSONKey: "BrandingThemes", IDField: "BrandingThemeID",
		Columns: []output.Column{
			{Header: "ID", Path: "BrandingThemeID"},
			{Header: "NAME", Path: "Name"},
			{Header: "SORT ORDER", Path: "SortOrder"},
			{Header: "CREATED", Path: "CreatedDateUTC", Format: "date"},
		},
		ReadOnly: true,
	}
	return cmdutil.NewResourceCmd(f, def)
}
