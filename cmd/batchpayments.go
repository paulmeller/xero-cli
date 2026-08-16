package cmd

import (
	"github.com/spf13/cobra"

	"github.com/paulmeller/xero-cli/internal/api"
	"github.com/paulmeller/xero-cli/internal/cmdutil"
	"github.com/paulmeller/xero-cli/internal/output"
)

func newBatchPaymentsCmd(f *cmdutil.Factory) *cobra.Command {
	def := cmdutil.ResourceDef{
		Name: "batch-payment", Plural: "batch-payments",
		APIPath: api.PathBatchPayments, JSONKey: "BatchPayments", IDField: "BatchPaymentID",
		Columns: []output.Column{
			{Header: "ID", Path: "BatchPaymentID"},
			{Header: "DATE", Path: "Date", Format: "date"},
			{Header: "ACCOUNT", Path: "Account.Name"},
			{Header: "REFERENCE", Path: "Reference"},
			{Header: "TOTAL", Path: "TotalAmount", Format: "currency"},
			{Header: "STATUS", Path: "Status", Format: "status"},
		},
		HasCreate: true, HasDelete: true,
		// Xero's BatchPayments collection is create-via-PUT; POST on this collection maps to
		// deleteBatchPayment, not update - routing create through POST silently deletes instead.
		CreateUsesPut:     true,
		PutBatchSupported: true, // "Creates one or many batch payments" per the OpenAPI spec
	}
	return cmdutil.NewResourceCmd(f, def)
}
