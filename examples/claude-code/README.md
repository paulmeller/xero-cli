# Claude Code Bookkeeping Setup

Use Claude Code as an AI bookkeeper with the `xero` CLI.

## Quick Start

1. **Install and authenticate the CLI:**

```bash
# Install
go install github.com/paulmeller/xero-cli@latest

# Authenticate (interactive)
xero auth login

# Or for CI/headless (client credentials):
export XERO_CLIENT_ID="..."
export XERO_CLIENT_SECRET="..."
export XERO_GRANT_TYPE="client_credentials"
```

2. **Copy this directory into your project:**

```bash
cp -r examples/claude-code/ ~/bookkeeping/
cd ~/bookkeeping/
```

3. **Customise for your org:**

Edit `.claude/skills/xero-bookkeeper/reference/workflows.md` with your
org-specific account codes, payment terms, and procedures.

Generate your chart of accounts reference:

```bash
echo "# Chart of Accounts (generated $(date -Iseconds))" \
  > .claude/skills/xero-bookkeeper/reference/chart-of-accounts.md
xero accounts list -o table \
  >> .claude/skills/xero-bookkeeper/reference/chart-of-accounts.md
```

Regenerate the CLI reference (after CLI updates):

```bash
echo "# xero CLI Reference" > .claude/skills/xero-bookkeeper/reference/cli-reference.md
xero --help >> .claude/skills/xero-bookkeeper/reference/cli-reference.md
for cmd in invoices contacts payments accounts bank-transactions \
           credit-notes items reports tax-rates tracking currencies \
           manual-journals purchase-orders quotes sync tenants org \
           journals overpayments prepayments batch-payments \
           linked-transactions repeating-invoices branding-themes \
           rate-limits; do
  echo -e "\n---\n## xero $cmd\n" >> .claude/skills/xero-bookkeeper/reference/cli-reference.md
  xero $cmd --help >> .claude/skills/xero-bookkeeper/reference/cli-reference.md 2>/dev/null
done
```

4. **Start Claude Code:**

```bash
cd ~/bookkeeping/
claude
```

5. **Use slash commands:**

```
> /overdue
> /invoice
> /month-end
> /reconcile
```

## Directory Structure

```
bookkeeping/
├── .claude/
│   ├── skills/
│   │   └── xero-bookkeeper/
│   │       ├── SKILL.md                  # Agent skill (auto-loaded)
│   │       └── reference/
│   │           ├── cli-reference.md      # Full CLI command reference
│   │           ├── chart-of-accounts.md  # Your org's accounts (generate)
│   │           └── workflows.md          # Your org's procedures (edit)
│   └── commands/
│       ├── month-end.md                  # /month-end
│       ├── overdue.md                    # /overdue
│       ├── invoice.md                    # /invoice
│       └── reconcile.md                  # /reconcile
├── CLAUDE.md                             # Project context
├── data/                                 # Temp JSON files for create/update
└── exports/                              # CSV/report output
```

## Safety Model

Three layers protect your financial data:

1. **CLI level** -- `--dry-run` previews any write without executing.
2. **Skill level** -- SKILL.md classifies commands as read (safe) vs write
   (confirm first) and encodes rules like "never void invoices with payments."
3. **Claude Code level** -- You see every command before it runs and can reject it.

For write operations, Claude will: explain the intent, run `--dry-run`, and
wait for your explicit approval before executing.
