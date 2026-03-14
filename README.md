# xero-cli

A command-line interface for the [Xero](https://www.xero.com/) accounting API.

```
xero invoices list
xero contacts get <id> -o json
xero payments create --invoice INV-001 --account 090 --amount 500 --date 2026-03-14
xero reports profit-and-loss --from-date 2026-01-01 --to-date 2026-03-31
xero sync
```

## Install

**Homebrew:**

```bash
brew tap paulmeller/tap
brew install xero-cli
```

**Go:**

```bash
go install github.com/paulmeller/xero-cli@latest
```

**From source:**

```bash
git clone https://github.com/paulmeller/xero-cli.git
cd xero-cli
make install
```

## Authenticate

Create a [Xero app](https://developer.xero.com/app/manage/) (Web App type), set the redirect URI to `http://localhost:8472/callback`, then:

```bash
# Interactive login (opens browser)
export XERO_CLIENT_ID="your-client-id"
export XERO_CLIENT_SECRET="your-client-secret"
xero auth login

# Or use a config file (~/.config/xero-cli/config.toml)
cat > ~/.config/xero-cli/config.toml << 'EOF'
client_id = "your-client-id"
client_secret = "your-client-secret"
EOF
xero auth login
```

For CI / headless environments:

```bash
export XERO_CLIENT_ID="..."
export XERO_CLIENT_SECRET="..."
export XERO_GRANT_TYPE="client_credentials"
xero auth login
```

Tokens are stored in the OS keychain (macOS Keychain, Windows Credential Manager, Linux Secret Service) with automatic fallback to `~/.config/xero-cli/tokens.json`.

## Commands

### Resources

| Command | Operations |
|---------|-----------|
| `invoices` | list, get, create, update, delete, void, email, online-url, history, attach |
| `contacts` | list, get, create, update, archive, history, attach |
| `payments` | list, get, create, delete |
| `accounts` | list, get, create, update, archive, attach |
| `bank-transactions` | list, get, create, update, history, attach |
| `credit-notes` | list, get, create, update, allocate |
| `manual-journals` | list, get, create, update |
| `purchase-orders` | list, get, create, update, history, attach |
| `items` | list, get, create, update, delete, history |
| `quotes` | list, get, create, update, history |
| `batch-payments` | list, get, create, delete |
| `overpayments` | list, get, allocate |
| `prepayments` | list, get, allocate |
| `linked-transactions` | list, get, create, update, delete |
| `repeating-invoices` | list, get, history, attach |
| `journals` | list, get |
| `tax-rates` | list, get, create, update |
| `currencies` | list, create |
| `branding-themes` | list |
| `tracking` | categories and options CRUD |

### Reports

```bash
xero reports profit-and-loss --from-date 2026-01-01 --to-date 2026-03-31
xero reports balance-sheet --date 2026-03-31
xero reports trial-balance --date 2026-03-31
xero reports aged-receivables --date 2026-03-31
xero reports aged-payables --date 2026-03-31
xero reports bank-summary --from-date 2026-01-01 --to-date 2026-03-31
xero reports budget-summary --date 2026-03-31
xero reports executive-summary --date 2026-03-31
```

### Other

```bash
xero auth login|logout|status|refresh|migrate-keychain
xero tenants list|switch|current
xero organisation info|actions
xero rate-limits
xero sync [--streams invoices,contacts] [--full-refresh] [--dry-run]
xero completion bash|zsh|fish
```

## Output Formats

```bash
xero invoices list              # table (default in terminal)
xero invoices list -o json      # JSON (default when piped)
xero invoices list -o csv       # CSV
xero invoices list -o tsv       # TSV (for awk/cut/sort)
```

TSV is pipe-friendly -- no quoting, no alignment:

```bash
xero invoices list -o tsv | cut -f2,3,6 | sort -t$'\t' -k3
```

## Filtering and Pagination

```bash
xero invoices list --where 'Status=="AUTHORISED"'
xero invoices list --where 'AmountDue>1000' --order 'DueDate ASC'
xero contacts list --search "Acme"
xero invoices list --modified-since 2026-03-01T00:00:00
xero invoices list --all                    # auto-paginate
xero invoices list --page 2 --page-size 50  # manual pagination
```

## Creating and Updating

```bash
# Inline flags
xero invoices create --contact "Acme Corp" --date 2026-03-14 \
  --due-date 2026-04-14 --line "Consulting,10,150,200"

# From JSON file
xero invoices create --file invoice.json

# Batch create (JSON array, up to 50 items)
xero invoices create --file invoices.json

# From stdin
echo '{"Name":"New Contact"}' | xero contacts create --file -

# Update
xero invoices update <id> --status AUTHORISED
xero contacts update <id> --file updated-contact.json
```

## Agent / Automation Flags

```bash
--no-prompt         # never prompt; fail if confirmation needed (auto-set when stdin is not a TTY)
--force             # skip confirmation for destructive actions
--dry-run           # print HTTP request without sending
--idempotency-key   # explicit key for retry safety (auto-generated if omitted)
-o json             # structured output
--file -            # read from stdin
```

## Sync (ELT)

Sync Xero data locally for fast querying without API calls:

```bash
xero sync init                              # generate sync.toml
xero sync                                   # incremental sync
xero sync --streams invoices,contacts       # specific streams
xero sync --full-refresh                    # ignore bookmarks
xero sync status                            # last sync times
xero sync reset [stream]                    # clear state
```

Destinations: JSONL files, DuckDB, or stdout. Configure in `sync.toml`:

```toml
[destination]
type = "jsonl"
output_dir = "./xero_data"

[[streams]]
name = "invoices"
enabled = true
sync_mode = "incremental"
```

## Claude Code Integration

Use xero-cli as an AI bookkeeper with Claude Code. See [`examples/claude-code/`](examples/claude-code/) for a ready-to-use setup with:

- Agent skill with safety rules and workflow procedures
- Slash commands: `/invoice`, `/overdue`, `/month-end`, `/reconcile`
- Three-layer safety model (CLI dry-run, skill rules, human approval)

## Configuration

`~/.config/xero-cli/config.toml`:

```toml
client_id = "..."
client_secret = "..."
active_tenant = "..."
scopes = ["openid", "offline_access", "accounting.invoices", "accounting.contacts", "..."]
redirect_uri = "http://localhost:8472/callback"

[defaults]
output = "table"
page_size = 100
```

Environment variables override config: `XERO_CLIENT_ID`, `XERO_CLIENT_SECRET`, `XERO_TENANT_ID`, `XERO_GRANT_TYPE`.

## License

MIT
