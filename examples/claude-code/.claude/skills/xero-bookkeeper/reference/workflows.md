# Org-Specific Workflows

Customise this file for your organisation. This is where you document
how YOUR business does things -- payment terms, approval chains,
account code conventions, month-end checklists, etc.

Claude reads this file to understand your specific procedures.

## Example Sections (replace with your own)

### Invoice Conventions
- Sales invoices use type ACCREC, account code 200 (Sales Revenue)
- All invoices must have a PO number in the Reference field
- Payment terms: Net 30 for domestic, Net 60 for international
- Tax: GST Inclusive for domestic, No Tax for international

### Payment Processing
- Bank account for receiving: "Business Cheque Account" (code 090)
- Bank account for paying: "Business Cheque Account" (code 090)
- Payments over $10,000 require manual approval before recording

### Month-End Checklist
1. Reconcile all bank accounts
2. Review and approve draft invoices
3. Check for unmatched bank transactions
4. Review aged receivables -- follow up on 30+ day overdue
5. Run trial balance and check suspense account (code 800) is zero
6. Generate P&L and Balance Sheet reports
7. Lock the period

### Account Code Quick Reference
- 200: Sales Revenue
- 400: Office Expenses
- 410: Travel & Entertainment
- 090: Business Cheque Account
- 800: Suspense (should always be zero)

### Client-Specific Notes
- Acme Corp: always reference their PO numbers
- Globex Inc: payment terms are Net 45 (special agreement)
