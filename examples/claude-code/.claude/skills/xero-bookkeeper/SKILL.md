---
name: xero-bookkeeper
description: >
  Bookkeeping assistant for Xero. Triggers on: invoices, payments, bank
  transactions, contacts, accounts, reports, reconciliation, month-end,
  tax prep, P&L, balance sheet, overdue, aged receivables/payables,
  trial balance, chart of accounts, or any mention of "Xero".
---

# Xero Bookkeeping Skill

You have access to the `xero` CLI for managing accounting data.
Run commands via Bash. Use `-o json` for data processing, `-o table`
for user-facing output.

Before starting, consult:
- `reference/cli-reference.md` for full command syntax
- `reference/chart-of-accounts.md` for this org's account codes
- `reference/workflows.md` for org-specific procedures

## Read Commands (safe -- run without asking)

```
xero invoices list [--status X] [--where "..."] [-o json]
xero invoices get <id> [-o json]
xero contacts list [--search "name"] [-o json]
xero contacts get <id> [-o json]
xero payments list [-o json]
xero accounts list [-o json]
xero bank-transactions list [-o json]
xero items list [-o json]
xero tax-rates list [-o json]
xero org info [-o json]
xero reports profit-and-loss [--from-date X --to-date X] [-o json]
xero reports balance-sheet [--date X] [-o json]
xero reports trial-balance [-o json]
xero reports aged-receivables [-o json]
xero reports aged-payables [-o json]
xero reports bank-summary [-o json]
xero tenants list [-o json]
xero sync status
xero rate-limits
```

## Write Commands (ALWAYS confirm with user first)

```
xero invoices create --file <f> [--dry-run]
xero invoices update <id> --status AUTHORISED
xero invoices void <id>
xero invoices delete <id>
xero invoices email <id>
xero contacts create --name "..." --email "..."
xero payments create --invoice <id> --account <code> --amount <n> --date <d>
xero payments delete <id>
xero credit-notes create --file <f>
xero manual-journals create --file <f>
xero sync
```

## Safety Rules

1. Read freely, write carefully. For ANY write command:
   a. Explain what you plan to do
   b. Run with --dry-run and show the output
   c. Wait for explicit "yes" before executing

2. Never modify locked periods. Check `xero org info -o json` for
   PeriodLockDate before writing to past dates.

3. Never void/delete invoices with payments. Check AmountPaid first.

4. Bulk operations (>5 records): show a summary table, get approval,
   process one at a time with progress updates.

5. Never guess account codes. Run `xero accounts list` and present
   options if unsure.

6. Check currency via `xero org info` -- don't assume.

## Core Workflows

### Create an Invoice
1. `xero contacts list --search "<name>" -o json` -- find contact
2. If ambiguous, show options and ask user to pick
3. Write invoice JSON to ./data/invoice.json
4. `xero invoices create --file ./data/invoice.json --dry-run` -- preview
5. On approval: `xero invoices create --file ./data/invoice.json -o json`
6. If asked: `xero invoices email <id>`

### Record a Payment
1. `xero invoices list --numbers "INV-XXX" -o json` -- find invoice
2. Show: contact, amount, status, due date
3. `xero accounts list --where "Type==\"BANK\"" -o json` -- get bank accounts
4. `xero payments create ... --dry-run` -- preview
5. On approval: execute

### Reconciliation
1. `xero bank-transactions list -o json --all` -- unreconciled txns
2. `xero invoices list --status AUTHORISED -o json --all` -- outstanding invoices
3. Match by: exact amount, date proximity, contact name
4. Present match table with confidence level
5. Create payments one at a time for confirmed matches

### Month-End Close
1. `xero sync` (ask permission -- this is a write)
2. `xero reports trial-balance -o json`
3. Check for: unreconciled bank txns, draft invoices, suspense balances
4. `xero reports profit-and-loss --from-date <start> --to-date <end> -o json`
5. `xero reports balance-sheet --date <end> -o json`
6. Present summary with flagged items

### Aged Analysis
1. `xero reports aged-receivables -o json` -- owed to us
2. `xero reports aged-payables -o json` -- we owe
3. Group by bucket (current / 30 / 60 / 90+ days)
4. Highlight largest and most overdue

### Multi-Tenant (Accounting Firms)
1. `xero tenants list -o json`
2. Loop with --tenant <id> per command
3. Present cross-client summary
4. Drill into specific tenant on request
