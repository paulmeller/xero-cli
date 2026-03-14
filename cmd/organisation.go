package cmd

import (
	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"

	"github.com/paulmeller/xero-cli/internal/api"
	"github.com/paulmeller/xero-cli/internal/cmdutil"
	"github.com/paulmeller/xero-cli/internal/output"
)

func newOrganisationCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "organisation",
		Aliases: []string{"org", "organization"},
		Short:   "Organisation information",
	}

	cmd.AddCommand(newOrgInfoCmd(f))
	cmd.AddCommand(newOrgActionsCmd(f))
	return cmd
}

func newOrgActionsCmd(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "actions",
		Short: "List available organisation actions",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}
			cmdutil.ApplyClientFlags(cmd, client, f)

			data, err := client.Get(cmd.Context(), "Organisation/Actions", nil)
			if err != nil {
				return err
			}

			format := cmdutil.GetOutputFormat(cmd, f.IO)
			formatter := f.Formatter(format)
			return formatter.FormatOne(f.IO.Out, gjson.ParseBytes(data), nil)
		},
	}
}

func newOrgInfoCmd(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "info",
		Short: "Show organisation details",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}
			cmdutil.ApplyClientFlags(cmd, client, f)

			data, err := client.Get(cmd.Context(), api.PathOrganisation, nil)
			if err != nil {
				return err
			}

			format := cmdutil.GetOutputFormat(cmd, f.IO)
			formatter := f.Formatter(format)

			item := gjson.ParseBytes(data).Get("Organisations.0")
			if !item.Exists() {
				item = gjson.ParseBytes(data)
			}

			columns := []output.Column{
				{Header: "NAME", Path: "Name"},
				{Header: "LEGAL NAME", Path: "LegalName"},
				{Header: "SHORT CODE", Path: "ShortCode"},
				{Header: "ORG TYPE", Path: "OrganisationType"},
				{Header: "COUNTRY", Path: "CountryCode"},
				{Header: "CURRENCY", Path: "BaseCurrency"},
				{Header: "TAX NUMBER", Path: "TaxNumber"},
				{Header: "FINANCIAL YEAR END", Path: "FinancialYearEndDay"},
				{Header: "SALES TAX BASIS", Path: "SalesTaxBasis"},
			}

			return formatter.FormatOne(f.IO.Out, item, columns)
		},
	}
}
