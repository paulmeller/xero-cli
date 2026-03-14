package cmd

import (
	"github.com/spf13/cobra"

	"github.com/paulmeller/xero-cli/internal/api"
	"github.com/paulmeller/xero-cli/internal/cmdutil"
	"github.com/paulmeller/xero-cli/internal/output"
)

func newItemsCmd(f *cmdutil.Factory) *cobra.Command {
	def := cmdutil.ResourceDef{
		Name: "item", Plural: "items",
		APIPath: api.PathItems, JSONKey: "Items", IDField: "ItemID",
		Columns: []output.Column{
			{Header: "ID", Path: "ItemID"},
			{Header: "CODE", Path: "Code"},
			{Header: "NAME", Path: "Name"},
			{Header: "DESCRIPTION", Path: "Description"},
			{Header: "PURCHASE PRICE", Path: "PurchaseDetails.UnitPrice", Format: "currency"},
			{Header: "SALE PRICE", Path: "SalesDetails.UnitPrice", Format: "currency"},
		},
		HasCreate: true, HasUpdate: true, HasDelete: true, HasHistory: true,
	}
	return cmdutil.NewResourceCmd(f, def)
}
