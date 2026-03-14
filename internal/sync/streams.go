package sync

// StreamMeta maps a stream name to its API path and JSON key.
type StreamMeta struct {
	APIPath string
	JSONKey string
}

// Registry of all known streams.
var StreamRegistry = map[string]StreamMeta{
	"invoices":            {APIPath: "Invoices", JSONKey: "Invoices"},
	"contacts":            {APIPath: "Contacts", JSONKey: "Contacts"},
	"accounts":            {APIPath: "Accounts", JSONKey: "Accounts"},
	"payments":            {APIPath: "Payments", JSONKey: "Payments"},
	"bank_transactions":   {APIPath: "BankTransactions", JSONKey: "BankTransactions"},
	"credit_notes":        {APIPath: "CreditNotes", JSONKey: "CreditNotes"},
	"manual_journals":     {APIPath: "ManualJournals", JSONKey: "ManualJournals"},
	"purchase_orders":     {APIPath: "PurchaseOrders", JSONKey: "PurchaseOrders"},
	"items":               {APIPath: "Items", JSONKey: "Items"},
	"journals":            {APIPath: "Journals", JSONKey: "Journals"},
	"quotes":              {APIPath: "Quotes", JSONKey: "Quotes"},
	"tax_rates":           {APIPath: "TaxRates", JSONKey: "TaxRates"},
	"tracking_categories": {APIPath: "TrackingCategories", JSONKey: "TrackingCategories"},
	"currencies":          {APIPath: "Currencies", JSONKey: "Currencies"},
	"organisation":        {APIPath: "Organisation", JSONKey: "Organisations"},
	"branding_themes":     {APIPath: "BrandingThemes", JSONKey: "BrandingThemes"},
	"overpayments":        {APIPath: "Overpayments", JSONKey: "Overpayments"},
	"prepayments":         {APIPath: "Prepayments", JSONKey: "Prepayments"},
	"repeating_invoices":  {APIPath: "RepeatingInvoices", JSONKey: "RepeatingInvoices"},
	"batch_payments":      {APIPath: "BatchPayments", JSONKey: "BatchPayments"},
	"linked_transactions": {APIPath: "LinkedTransactions", JSONKey: "LinkedTransactions"},
}

// StreamPriority defines sync order (high-change streams first).
var StreamPriority = []string{
	"invoices",
	"bank_transactions",
	"payments",
	"contacts",
	"credit_notes",
	"manual_journals",
	"purchase_orders",
	"items",
	"journals",
	"quotes",
	"overpayments",
	"prepayments",
	"batch_payments",
	"linked_transactions",
	"accounts",
	"tax_rates",
	"tracking_categories",
	"currencies",
	"organisation",
	"branding_themes",
	"repeating_invoices",
}
