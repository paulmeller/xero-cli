package sync

// StreamMeta maps a stream name to its API path, JSON key, and primary key.
type StreamMeta struct {
	APIPath    string
	JSONKey    string
	PrimaryKey string
}

// Registry of all known streams.
var StreamRegistry = map[string]StreamMeta{
	"invoices":            {APIPath: "Invoices", JSONKey: "Invoices", PrimaryKey: "InvoiceID"},
	"contacts":            {APIPath: "Contacts", JSONKey: "Contacts", PrimaryKey: "ContactID"},
	"accounts":            {APIPath: "Accounts", JSONKey: "Accounts", PrimaryKey: "AccountID"},
	"payments":            {APIPath: "Payments", JSONKey: "Payments", PrimaryKey: "PaymentID"},
	"bank_transactions":   {APIPath: "BankTransactions", JSONKey: "BankTransactions", PrimaryKey: "BankTransactionID"},
	"credit_notes":        {APIPath: "CreditNotes", JSONKey: "CreditNotes", PrimaryKey: "CreditNoteID"},
	"manual_journals":     {APIPath: "ManualJournals", JSONKey: "ManualJournals", PrimaryKey: "ManualJournalID"},
	"purchase_orders":     {APIPath: "PurchaseOrders", JSONKey: "PurchaseOrders", PrimaryKey: "PurchaseOrderID"},
	"items":               {APIPath: "Items", JSONKey: "Items", PrimaryKey: "ItemID"},
	"journals":            {APIPath: "Journals", JSONKey: "Journals", PrimaryKey: "JournalID"},
	"quotes":              {APIPath: "Quotes", JSONKey: "Quotes", PrimaryKey: "QuoteID"},
	"tax_rates":           {APIPath: "TaxRates", JSONKey: "TaxRates", PrimaryKey: "TaxType"},
	"tracking_categories": {APIPath: "TrackingCategories", JSONKey: "TrackingCategories", PrimaryKey: "TrackingCategoryID"},
	"currencies":          {APIPath: "Currencies", JSONKey: "Currencies", PrimaryKey: "Code"},
	"organisation":        {APIPath: "Organisation", JSONKey: "Organisations", PrimaryKey: "OrganisationID"},
	"branding_themes":     {APIPath: "BrandingThemes", JSONKey: "BrandingThemes", PrimaryKey: "BrandingThemeID"},
	"overpayments":        {APIPath: "Overpayments", JSONKey: "Overpayments", PrimaryKey: "OverpaymentID"},
	"prepayments":         {APIPath: "Prepayments", JSONKey: "Prepayments", PrimaryKey: "PrepaymentID"},
	"repeating_invoices":  {APIPath: "RepeatingInvoices", JSONKey: "RepeatingInvoices", PrimaryKey: "RepeatingInvoiceID"},
	"batch_payments":      {APIPath: "BatchPayments", JSONKey: "BatchPayments", PrimaryKey: "BatchPaymentID"},
	"linked_transactions": {APIPath: "LinkedTransactions", JSONKey: "LinkedTransactions", PrimaryKey: "LinkedTransactionID"},
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
