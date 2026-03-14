package cmd

import (
	"github.com/spf13/cobra"

	"github.com/paulmeller/xero-cli/internal/api"
	"github.com/paulmeller/xero-cli/internal/cmdutil"
	"github.com/paulmeller/xero-cli/internal/output"
)

func newPurchaseOrdersCmd(f *cmdutil.Factory) *cobra.Command {
	def := cmdutil.ResourceDef{
		Name: "purchase-order", Plural: "purchase-orders",
		APIPath: api.PathPurchaseOrders, JSONKey: "PurchaseOrders", IDField: "PurchaseOrderID",
		Columns: []output.Column{
			{Header: "ID", Path: "PurchaseOrderID"},
			{Header: "NUMBER", Path: "PurchaseOrderNumber"},
			{Header: "CONTACT", Path: "Contact.Name"},
			{Header: "DATE", Path: "Date", Format: "date"},
			{Header: "STATUS", Path: "Status", Format: "status"},
			{Header: "TOTAL", Path: "Total", Format: "currency"},
		},
		HasCreate: true, HasUpdate: true, HasHistory: true, HasAttach: true,
	}
	return cmdutil.NewResourceCmd(f, def)
}
