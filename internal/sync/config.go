package sync

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

type SyncConfig struct {
	Destination DestinationConfig `toml:"destination"`
	Sync        SyncSettings      `toml:"sync"`
	Streams     []StreamConfig    `toml:"streams"`
}

type DestinationConfig struct {
	Type             string `toml:"type"`
	OutputDir        string `toml:"output_dir"`
	ConnectionString string `toml:"connection_string"`
}

type SyncSettings struct {
	Mode              string `toml:"mode"`
	StateFile         string `toml:"state_file"`
	Concurrency       int    `toml:"concurrency"`
	RequestsPerMinute int    `toml:"requests_per_minute"`
	DailyBudget       int    `toml:"daily_budget"`
}

type StreamConfig struct {
	Name             string   `toml:"name"`
	Enabled          bool     `toml:"enabled"`
	SyncMode         string   `toml:"sync_mode"`
	CursorField      string   `toml:"cursor_field"`
	PrimaryKey       string   `toml:"primary_key"`
	DestinationTable string   `toml:"destination_table"`
	SelectedFields   []string `toml:"selected_fields"`
	Where            string   `toml:"where"`
}

func LoadSyncConfig(path string) (*SyncConfig, error) {
	cfg := &SyncConfig{
		Destination: DestinationConfig{
			Type:      "jsonl",
			OutputDir: "./xero_data",
		},
		Sync: SyncSettings{
			Mode:              "incremental",
			StateFile:         ".xero-sync-state.json",
			Concurrency:       3,
			RequestsPerMinute: 55,
			DailyBudget:       4500,
		},
	}

	if path == "" {
		path = "sync.toml"
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("sync config not found: %s\nRun 'xero sync init' to create one", path)
	}

	if _, err := toml.DecodeFile(path, cfg); err != nil {
		return nil, fmt.Errorf("cannot parse sync config %s: %w", path, err)
	}

	return cfg, nil
}

func DefaultSyncConfig() string {
	return `# xero-cli sync configuration

[destination]
type = "jsonl"                          # jsonl | duckdb | stdout
output_dir = "./xero_data"
# connection_string = "./xero.duckdb"   # for duckdb

[sync]
mode = "incremental"
state_file = ".xero-sync-state.json"
concurrency = 3
requests_per_minute = 55
daily_budget = 4500

[[streams]]
name = "invoices"
enabled = true
sync_mode = "incremental"
cursor_field = "UpdatedDateUTC"
primary_key = "InvoiceID"

[[streams]]
name = "contacts"
enabled = true
sync_mode = "incremental"
cursor_field = "UpdatedDateUTC"
primary_key = "ContactID"

[[streams]]
name = "accounts"
enabled = true
sync_mode = "full_refresh"
cursor_field = ""
primary_key = "AccountID"

[[streams]]
name = "payments"
enabled = true
sync_mode = "incremental"
cursor_field = "UpdatedDateUTC"
primary_key = "PaymentID"

[[streams]]
name = "bank_transactions"
enabled = true
sync_mode = "incremental"
cursor_field = "UpdatedDateUTC"
primary_key = "BankTransactionID"

[[streams]]
name = "credit_notes"
enabled = true
sync_mode = "incremental"
cursor_field = "UpdatedDateUTC"
primary_key = "CreditNoteID"

[[streams]]
name = "manual_journals"
enabled = true
sync_mode = "incremental"
cursor_field = "UpdatedDateUTC"
primary_key = "ManualJournalID"

[[streams]]
name = "purchase_orders"
enabled = true
sync_mode = "incremental"
cursor_field = "UpdatedDateUTC"
primary_key = "PurchaseOrderID"

[[streams]]
name = "items"
enabled = true
sync_mode = "incremental"
cursor_field = "UpdatedDateUTC"
primary_key = "ItemID"

[[streams]]
name = "journals"
enabled = true
sync_mode = "incremental"
cursor_field = "CreatedDateUTC"
primary_key = "JournalID"

[[streams]]
name = "quotes"
enabled = true
sync_mode = "incremental"
cursor_field = "UpdatedDateUTC"
primary_key = "QuoteID"

[[streams]]
name = "tax_rates"
enabled = true
sync_mode = "full_refresh"
cursor_field = ""
primary_key = "TaxType"

[[streams]]
name = "tracking_categories"
enabled = true
sync_mode = "full_refresh"
cursor_field = ""
primary_key = "TrackingCategoryID"

[[streams]]
name = "currencies"
enabled = true
sync_mode = "full_refresh"
cursor_field = ""
primary_key = "Code"

[[streams]]
name = "organisation"
enabled = true
sync_mode = "full_refresh"
cursor_field = ""
primary_key = "OrganisationID"

[[streams]]
name = "branding_themes"
enabled = true
sync_mode = "full_refresh"
cursor_field = ""
primary_key = "BrandingThemeID"

[[streams]]
name = "overpayments"
enabled = true
sync_mode = "incremental"
cursor_field = "UpdatedDateUTC"
primary_key = "OverpaymentID"

[[streams]]
name = "prepayments"
enabled = true
sync_mode = "incremental"
cursor_field = "UpdatedDateUTC"
primary_key = "PrepaymentID"

[[streams]]
name = "repeating_invoices"
enabled = true
sync_mode = "full_refresh"
cursor_field = ""
primary_key = "RepeatingInvoiceID"

[[streams]]
name = "batch_payments"
enabled = true
sync_mode = "incremental"
cursor_field = "UpdatedDateUTC"
primary_key = "BatchPaymentID"

[[streams]]
name = "linked_transactions"
enabled = true
sync_mode = "incremental"
cursor_field = "UpdatedDateUTC"
primary_key = "LinkedTransactionID"
`
}
