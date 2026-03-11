# relay ERP Roadmap

Suggested evolution from relay's current state toward a full-featured ERP,
ordered by business value and technical dependency.

---

## Phase 0 — Current State ✅

- Companies (tenants)
- Users (authentication stub)
- Transactions (simplified debit/credit ledger)

---

## Phase 1 — Financial Foundation

**Goal**: be a real financial management tool.

### 1.1 Chart of Accounts
- `accounts` table with code, name, type (asset/liability/equity/revenue/expense)
- Per-company chart of accounts
- Seed default chart of accounts on company creation

### 1.2 Upgrade Transactions → GL Posting
- Add `account_id` FK to `transactions`
- Add `fiscal_period` (YYYY-MM) field
- Add `status` (draft/posted/void) with lifecycle enforcement in controller
- Add `reference` field for linking to invoices/payments

### 1.3 Fiscal Periods
- `fiscal_periods` table: `year`, `month`, `status` (open/closed)
- Enforce: cannot post to a closed period
- Month-end close workflow

**Deliverable**: company can post accounting entries to specific GL accounts and periods.

---

## Phase 2 — Accounts Payable

**Goal**: track and pay vendor bills.

### 2.1 Vendors
- `vendors` table: name, CNPJ, email, payment terms, bank details
- CRUD + search by company

### 2.2 AP Invoices
- `invoices` table with `vendor_id`, `number`, `invoice_date`, `due_date`, `amount`
- Status machine: draft → posted → paid → void
- On post: auto-generate journal entry (debit expense account, credit AP account)

### 2.3 Payments (AP)
- `payments` table linked to `invoices`
- On payment: auto-generate journal entry (debit AP account, credit bank account)
- Mark invoice as paid

**Deliverable**: company can manage the full purchase-to-pay cycle.

---

## Phase 3 — Accounts Receivable

**Goal**: track and collect customer payments.

### 3.1 Customers
- `customers` table: name, CPF/CNPJ, email, credit limit, payment terms
- Credit limit enforcement in invoice creation

### 3.2 AR Invoices (Sales Invoices)
- Reuse `invoices` table with `customer_id` (polymorphic via type, or separate table)
- Status machine same as AP
- On post: auto-generate journal entry (debit AR account, credit revenue account)

### 3.3 Payment Receipts (AR)
- Record payment received from customer
- On receipt: debit bank account, credit AR account

### 3.4 Aging Report
- AP aging: what we owe, how overdue
- AR aging: what customers owe us, how overdue

**Deliverable**: company can manage the order-to-cash cycle.

---

## Phase 4 — Cash Management

**Goal**: reconcile books with bank.

### 4.1 Bank Accounts
- `bank_accounts` table: bank name, branch, account number, current balance

### 4.2 Bank Transactions Import
- Import bank statement (CSV or OFX)
- Auto-match to posted transactions

### 4.3 Reconciliation
- Mark transactions as reconciled
- Show unreconciled items

**Deliverable**: company can close the books each month with confidence.

---

## Phase 5 — Reporting

**Goal**: financial statements and management reports.

### 5.1 Balance Sheet
- Assets = Liabilities + Equity (at a point in time)
- Derived from GL account balances

### 5.2 Income Statement (P&L)
- Revenue – Expenses = Net Income (for a period)

### 5.3 Cash Flow Statement
- Operating, investing, financing activities

### 5.4 Management Reports
- AP aging, AR aging
- Cash position forecast
- Expense by category

---

## Phase 6 — AI / Conversational Layer

relay's stated goal is to be "adaptable to AI and conversational flows."

### 6.1 Natural Language Queries
- "Show me all unpaid invoices from last month"
- "What is our cash position today?"
- "Which vendors are we most overdue with?"

### 6.2 Automated Categorization
- AI suggests GL account when a transaction is entered
- Learns from past categorizations per company

### 6.3 Anomaly Detection
- Flag unusual transactions (amount outlier, duplicate invoice, etc.)

---

## Dependency Graph

```
Companies + Users (done)
    ↓
Chart of Accounts
    ↓
GL Transactions (upgrade current)
    ↓           ↓
Vendors       Customers
    ↓           ↓
AP Invoices   AR Invoices
    ↓           ↓
AP Payments   Payment Receipts
         ↓
    Bank Accounts
         ↓
    Reconciliation
         ↓
    Financial Statements
```
