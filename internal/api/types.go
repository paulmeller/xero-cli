package api

// Invoice represents an invoice for create/update operations.
type Invoice struct {
	Type            string     `json:"Type,omitempty"`
	Contact         *Contact   `json:"Contact,omitempty"`
	Date            string     `json:"Date,omitempty"`
	DueDate         string     `json:"DueDate,omitempty"`
	LineAmountTypes string     `json:"LineAmountTypes,omitempty"`
	InvoiceNumber   string     `json:"InvoiceNumber,omitempty"`
	Reference       string     `json:"Reference,omitempty"`
	Status          string     `json:"Status,omitempty"`
	CurrencyCode    string     `json:"CurrencyCode,omitempty"`
	LineItems       []LineItem `json:"LineItems,omitempty"`
	URL             string     `json:"Url,omitempty"`
}

// LineItem represents a line item on an invoice, credit note, etc.
type LineItem struct {
	Description string  `json:"Description,omitempty"`
	Quantity    float64 `json:"Quantity,omitempty"`
	UnitAmount  float64 `json:"UnitAmount,omitempty"`
	AccountCode string  `json:"AccountCode,omitempty"`
	TaxType     string  `json:"TaxType,omitempty"`
	ItemCode    string  `json:"ItemCode,omitempty"`
	LineItemID  string  `json:"LineItemID,omitempty"`
}

// Contact represents a contact for create/update operations.
type Contact struct {
	ContactID     string    `json:"ContactID,omitempty"`
	Name          string    `json:"Name,omitempty"`
	EmailAddress  string    `json:"EmailAddress,omitempty"`
	FirstName     string    `json:"FirstName,omitempty"`
	LastName      string    `json:"LastName,omitempty"`
	TaxNumber     string    `json:"TaxNumber,omitempty"`
	AccountNumber string    `json:"AccountNumber,omitempty"`
	ContactStatus string    `json:"ContactStatus,omitempty"`
	Phones        []Phone   `json:"Phones,omitempty"`
	Addresses     []Address `json:"Addresses,omitempty"`
}

// Phone represents a phone number on a contact.
type Phone struct {
	PhoneType        string `json:"PhoneType,omitempty"`
	PhoneNumber      string `json:"PhoneNumber,omitempty"`
	PhoneAreaCode    string `json:"PhoneAreaCode,omitempty"`
	PhoneCountryCode string `json:"PhoneCountryCode,omitempty"`
}

// Address represents an address on a contact.
type Address struct {
	AddressType  string `json:"AddressType,omitempty"`
	AddressLine1 string `json:"AddressLine1,omitempty"`
	AddressLine2 string `json:"AddressLine2,omitempty"`
	City         string `json:"City,omitempty"`
	Region       string `json:"Region,omitempty"`
	PostalCode   string `json:"PostalCode,omitempty"`
	Country      string `json:"Country,omitempty"`
}

// Payment represents a payment for create operations.
type Payment struct {
	Invoice   *Invoice `json:"Invoice,omitempty"`
	Account   *Account `json:"Account,omitempty"`
	Amount    float64  `json:"Amount,omitempty"`
	Date      string   `json:"Date,omitempty"`
	Reference string   `json:"Reference,omitempty"`
	Status    string   `json:"Status,omitempty"`
}

// Account represents a chart of accounts entry.
type Account struct {
	AccountID   string `json:"AccountID,omitempty"`
	Code        string `json:"Code,omitempty"`
	Name        string `json:"Name,omitempty"`
	Type        string `json:"Type,omitempty"`
	Description string `json:"Description,omitempty"`
	TaxType     string `json:"TaxType,omitempty"`
	Status      string `json:"Status,omitempty"`
	Class       string `json:"Class,omitempty"`
}

// CreditNote represents a credit note.
type CreditNote struct {
	Type             string     `json:"Type,omitempty"`
	Contact          *Contact   `json:"Contact,omitempty"`
	Date             string     `json:"Date,omitempty"`
	Status           string     `json:"Status,omitempty"`
	LineAmountTypes  string     `json:"LineAmountTypes,omitempty"`
	LineItems        []LineItem `json:"LineItems,omitempty"`
	CreditNoteNumber string    `json:"CreditNoteNumber,omitempty"`
	Reference        string    `json:"Reference,omitempty"`
}

// BankTransaction represents a bank transaction.
type BankTransaction struct {
	Type            string     `json:"Type,omitempty"`
	Contact         *Contact   `json:"Contact,omitempty"`
	BankAccount     *Account   `json:"BankAccount,omitempty"`
	Date            string     `json:"Date,omitempty"`
	Status          string     `json:"Status,omitempty"`
	LineAmountTypes string     `json:"LineAmountTypes,omitempty"`
	LineItems       []LineItem `json:"LineItems,omitempty"`
	Reference       string     `json:"Reference,omitempty"`
	IsReconciled    bool       `json:"IsReconciled,omitempty"`
}

// PurchaseOrder represents a purchase order.
type PurchaseOrder struct {
	Contact             *Contact   `json:"Contact,omitempty"`
	Date                string     `json:"Date,omitempty"`
	DeliveryDate        string     `json:"DeliveryDate,omitempty"`
	LineAmountTypes     string     `json:"LineAmountTypes,omitempty"`
	PurchaseOrderNumber string     `json:"PurchaseOrderNumber,omitempty"`
	Reference           string     `json:"Reference,omitempty"`
	Status              string     `json:"Status,omitempty"`
	LineItems           []LineItem `json:"LineItems,omitempty"`
}

// Item represents an inventory item.
type Item struct {
	Code        string `json:"Code,omitempty"`
	Name        string `json:"Name,omitempty"`
	Description string `json:"Description,omitempty"`
	IsSold      bool   `json:"IsSold,omitempty"`
	IsPurchased bool   `json:"IsPurchased,omitempty"`
}

// ManualJournal represents a manual journal entry.
type ManualJournal struct {
	Narration    string              `json:"Narration,omitempty"`
	Status       string              `json:"Status,omitempty"`
	Date         string              `json:"Date,omitempty"`
	JournalLines []ManualJournalLine `json:"JournalLines,omitempty"`
}

// ManualJournalLine represents a line in a manual journal.
type ManualJournalLine struct {
	LineAmount  float64 `json:"LineAmount,omitempty"`
	AccountCode string  `json:"AccountCode,omitempty"`
	Description string  `json:"Description,omitempty"`
	TaxType     string  `json:"TaxType,omitempty"`
}

// Allocation represents a credit note or prepayment allocation.
type Allocation struct {
	Invoice *Invoice `json:"Invoice,omitempty"`
	Amount  float64  `json:"Amount,omitempty"`
	Date    string   `json:"Date,omitempty"`
}

// Quote represents a quote.
type Quote struct {
	Contact         *Contact   `json:"Contact,omitempty"`
	Date            string     `json:"Date,omitempty"`
	ExpiryDate      string     `json:"ExpiryDate,omitempty"`
	Status          string     `json:"Status,omitempty"`
	LineAmountTypes string     `json:"LineAmountTypes,omitempty"`
	LineItems       []LineItem `json:"LineItems,omitempty"`
	Title           string     `json:"Title,omitempty"`
	Summary         string     `json:"Summary,omitempty"`
	Reference       string     `json:"Reference,omitempty"`
	Terms           string     `json:"Terms,omitempty"`
	QuoteNumber     string     `json:"QuoteNumber,omitempty"`
}

// TrackingCategory represents a tracking category.
type TrackingCategory struct {
	Name    string           `json:"Name,omitempty"`
	Status  string           `json:"Status,omitempty"`
	Options []TrackingOption `json:"Options,omitempty"`
}

// TrackingOption represents an option within a tracking category.
type TrackingOption struct {
	TrackingOptionID string `json:"TrackingOptionID,omitempty"`
	Name             string `json:"Name,omitempty"`
	Status           string `json:"Status,omitempty"`
}

// TaxRate represents a tax rate.
type TaxRate struct {
	Name          string `json:"Name,omitempty"`
	TaxType       string `json:"TaxType,omitempty"`
	ReportTaxType string `json:"ReportTaxType,omitempty"`
}

// Currency represents a currency.
type Currency struct {
	Code        string `json:"Code,omitempty"`
	Description string `json:"Description,omitempty"`
}

// LinkedTransaction represents a linked transaction.
type LinkedTransaction struct {
	SourceTransactionID string `json:"SourceTransactionID,omitempty"`
	SourceLineItemID    string `json:"SourceLineItemID,omitempty"`
	ContactID           string `json:"ContactID,omitempty"`
	TargetTransactionID string `json:"TargetTransactionID,omitempty"`
	TargetLineItemID    string `json:"TargetLineItemID,omitempty"`
	Status              string `json:"Status,omitempty"`
}

// BatchPayment represents a batch payment.
type BatchPayment struct {
	Account   *Account  `json:"Account,omitempty"`
	Date      string    `json:"Date,omitempty"`
	Payments  []Payment `json:"Payments,omitempty"`
	Reference string    `json:"Reference,omitempty"`
}

// HistoryRecord represents a history/note entry.
type HistoryRecord struct {
	Details string `json:"Details,omitempty"`
}
