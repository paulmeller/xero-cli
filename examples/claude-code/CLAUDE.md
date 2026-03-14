# Xero Bookkeeping Workspace

This project uses the `xero` CLI to manage bookkeeping for [Org Name].
The xero-bookkeeper skill in `.claude/skills/` has full CLI reference and
workflow procedures.

## Environment
- `xero` CLI is authenticated (env vars or ~/.config/xero-cli/tokens.json)
- Working directory: ./data/ for temp files, ./exports/ for outputs
- All amounts are in [NZD/AUD/GBP/USD] unless stated otherwise

## Slash Commands
- `/month-end` -- Run month-end close procedure
- `/overdue` -- Show overdue invoices ranked by amount
- `/invoice` -- Create and send an invoice interactively
- `/reconcile` -- Match bank transactions to invoices
