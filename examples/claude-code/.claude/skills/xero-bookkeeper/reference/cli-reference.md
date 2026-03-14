# xero CLI Reference

This file is auto-generated. Regenerate with:

```bash
echo "# xero CLI Reference" > cli-reference.md
xero --help >> cli-reference.md
for cmd in invoices contacts payments accounts bank-transactions \
           credit-notes items reports tax-rates tracking currencies \
           manual-journals purchase-orders quotes sync tenants org \
           journals overpayments prepayments batch-payments \
           linked-transactions repeating-invoices branding-themes \
           rate-limits; do
  echo -e "\n---\n## xero $cmd\n" >> cli-reference.md
  xero $cmd --help >> cli-reference.md 2>/dev/null
done
```

## Global Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--output` | `-o` | Output format: `table`, `json`, `csv` |
| `--tenant` | `-t` | Override active tenant for this call |
| `--quiet` | `-q` | Suppress non-essential output |
| `--verbose` | `-v` | Show HTTP request/response details |
| `--no-color` | | Disable colored output |
| `--no-prompt` | | Never prompt for confirmation; fail if required |
| `--page` | | Page number for paginated endpoints |
| `--page-size` | | Items per page (default: 100) |
| `--modified-since` | | Only return items modified after this ISO 8601 datetime |
| `--where` | | Xero API filter expression |
| `--order` | | Sort expression |
| `--dry-run` | | Show what would be sent without making the API call |
| `--config` | | Path to config file |

## Resources

### Invoices
```
xero invoices list [--status X] [--contact-id X] [--date-from X] [--date-to X] [--summary] [--numbers X] [--ids X] [--all]
xero invoices get <id>
xero invoices create --file <f> | --contact "Name" --line "Desc,Qty,Price,Account" [--type ACCREC|ACCPAY] [--date X] [--due-date X]
xero invoices update <id> --file <f> | --status X | --reference X | --due-date X
xero invoices delete <id> [--force]
xero invoices void <id> [--force]
xero invoices email <id>
xero invoices online-url <id>
xero invoices history <id>
xero invoices attach <id> <file-path>
```

### Contacts
```
xero contacts list [--search X] [--ids X] [--is-customer] [--is-supplier] [--include-archived] [--summary] [--all]
xero contacts get <id>
xero contacts create --file <f> | --name "Name" [--email X] [--phone X] [--tax-number X]
xero contacts update <id> --file <f>
xero contacts archive <id> [--force]
xero contacts history <id>
xero contacts attach <id> <file-path>
```

### Payments
```
xero payments list [--all]
xero payments get <id>
xero payments create --file <f> | --invoice <id> --account <code> --amount <n> --date <d>
xero payments delete <id> [--force]
```

### Accounts
```
xero accounts list [--all]
xero accounts get <id>
xero accounts create --file <f>
xero accounts update <id> --file <f>
xero accounts delete <id> [--force]
xero accounts archive <id> [--force]
xero accounts attach <id> <file-path>
```

### Bank Transactions
```
xero bank-transactions list [--all]
xero bank-transactions get <id>
xero bank-transactions create --file <f>
xero bank-transactions update <id> --file <f>
xero bank-transactions history <id>
xero bank-transactions attach <id> <file-path>
```

### Credit Notes
```
xero credit-notes list [--all]
xero credit-notes get <id>
xero credit-notes create --file <f>
xero credit-notes update <id> --file <f>
xero credit-notes allocate <id> --invoice <inv-id> --amount <n>
xero credit-notes attach <id> <file-path>
```

### Manual Journals
```
xero manual-journals list [--all]
xero manual-journals get <id>
xero manual-journals create --file <f>
xero manual-journals update <id> --file <f>
```

### Journals (read-only)
```
xero journals list [--all]
xero journals get <id>
```

### Reports
```
xero reports profit-and-loss [--from-date X] [--to-date X] [--periods N] [--timeframe X] [--tracking-category-id X] [--tracking-option-id X] [--payments-only]
xero reports balance-sheet [--date X] [--periods N] [--timeframe X] [--tracking-category-id X] [--tracking-option-id X] [--payments-only]
xero reports trial-balance [--date X] [--payments-only]
xero reports aged-receivables [--date X] [--from-date X] [--to-date X] [--periods N] [--timeframe X]
xero reports aged-payables [--date X] [--from-date X] [--to-date X] [--periods N] [--timeframe X]
xero reports bank-summary [--from-date X] [--to-date X]
xero reports budget-summary [--from-date X] [--to-date X] [--periods N] [--timeframe X]
xero reports executive-summary [--date X]
xero reports gst [--from-date X] [--to-date X]
xero reports 1099 [--from-date X] [--to-date X]
```

### Purchase Orders
```
xero purchase-orders list [--all]
xero purchase-orders get <id>
xero purchase-orders create --file <f>
xero purchase-orders update <id> --file <f>
xero purchase-orders history <id>
xero purchase-orders attach <id> <file-path>
```

### Items
```
xero items list [--all]
xero items get <id>
xero items create --file <f>
xero items update <id> --file <f>
xero items delete <id> [--force]
xero items history <id>
```

### Quotes
```
xero quotes list [--all]
xero quotes get <id>
xero quotes create --file <f>
xero quotes update <id> --file <f>
xero quotes history <id>
```

### Tracking Categories
```
xero tracking list
xero tracking get <category-id>
xero tracking create --name "Name"
xero tracking update <category-id> --name "New Name"
xero tracking delete <category-id> [--force]
xero tracking add-option <category-id> --name "Option"
xero tracking update-option <category-id> <option-id> --name "New Name"
xero tracking delete-option <category-id> <option-id> [--force]
```

### Other Resources
```
xero tax-rates list|get|create|update
xero currencies list|create
xero branding-themes list|get
xero batch-payments list|get|create|delete
xero linked-transactions list|get|create|update|delete
xero overpayments list|get|allocate
xero prepayments list|get|allocate
xero repeating-invoices list|get|history|attach
```

### Tenants
```
xero tenants list
xero tenants switch <tenant-id> | --name "Name"
xero tenants current
```

### Auth
```
xero auth login [--headless]
xero auth status
xero auth logout
xero auth refresh
```

### Sync
```
xero sync                              # Run sync using sync.toml
xero sync --streams invoices,contacts   # Sync specific streams
xero sync --full-refresh                # Ignore bookmarks, reload everything
xero sync --dry-run                     # Show what would sync
xero sync init                          # Generate default sync.toml
xero sync status                        # Show last sync times, row counts
xero sync reset [stream]                # Clear incremental state
```

### Rate Limits
```
xero rate-limits                        # Show current API rate limit status
```

### Organisation
```
xero org info [-o json]
xero org actions [-o json]
```
