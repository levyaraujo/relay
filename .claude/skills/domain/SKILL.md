---
name: domain
description: >
  ERP product domain knowledge for the relay project. Use this skill whenever
  designing, implementing, naming, or discussing ERP business concepts — financial
  modules, accounting flows, companies, users, transactions, invoices, chart of
  accounts, general ledger, accounts payable/receivable, cash management, fiscal
  periods, cost centers, or any feature that maps to standard ERP business logic.
  Triggers on: new ERP module, what should this domain contain, how does X work in
  ERP, design the data model for invoices/payments/GL/AP/AR, what fields does a
  transaction need, how do ERP financial flows work, implement ERP feature.
---

# ERP Domain Knowledge

Business domain reference for the **relay** ERP. Use this to make correct product
decisions — naming, data models, flows, validation rules — that align with how ERP
systems actually work in the market.

---

## The ERP Module Map

Every ERP is organized around business domains (modules). relay starts with
**Financial Management** as the core and expands outward.

```
Financial Management (core — build first)
├── General Ledger (GL)       ← single source of truth for all money
├── Accounts Payable (AP)     ← money owed TO vendors
├── Accounts Receivable (AR)  ← money owed BY customers
├── Cash Management           ← bank accounts, reconciliation, cash flow
└── Fixed Assets              ← depreciation, asset tracking

Operations (expand next)
├── Companies / Tenants       ← already exists in relay
├── Users & Permissions       ← already exists in relay
├── Customers (CRM-lite)      ← who owes us money
└── Vendors / Suppliers       ← who we owe money to

Reporting
├── Financial Statements      ← P&L, Balance Sheet, Cash Flow
├── Aging Reports             ← AP/AR aging
└── Period Close              ← month-end, year-end
```

---

## Core Financial Concepts

### General Ledger (GL)

The GL is the **single source of truth** for all financial data. Every financial
event — payment, invoice, expense — must post a journal entry to the GL.

Key entities:
- **Chart of Accounts** — numbered list of all accounts (assets, liabilities, equity, revenue, expenses)
- **Journal Entry** — a balanced debit/credit record with date, amount, description, accounts
- **Account** — one line in the chart of accounts (e.g. "1100 – Cash", "2000 – Accounts Payable")
- **Fiscal Period** — month or quarter; entries are posted to a period, periods are "closed" at month-end

A journal entry is always balanced: total debits = total credits.

### Transactions in relay

relay's current `transactions` table models a simplified ledger entry. It has:
- `type` (debit/credit) — direction of money flow
- `amount` — value
- `due_date` / `paid_date` — payment lifecycle
- `origin` — source system or human label
- `company_id` — tenant isolation (which company this belongs to)
- `creator_id` — who created it (audit trail)

**What's missing to reach full ERP financial capability:**
- `account_id` → which GL account this posts to
- `counterpart_account_id` → the offsetting account (for double-entry)
- `invoice_id` / `vendor_id` / `customer_id` → linkage to AP/AR
- `fiscal_period` → which accounting period
- `status` (draft, posted, void) → transaction lifecycle
- `currency` + `exchange_rate` → multi-currency support

### Accounts Payable (AP)

AP tracks money the company **owes to vendors**.

Core flow:
```
Purchase Order → Goods Receipt → Vendor Invoice → 3-Way Match → Payment → Reconciliation
```

Key entities:
- **Vendor** — the supplier (name, CNPJ/tax ID, payment terms, bank details)
- **Purchase Invoice** — bill received from a vendor (invoice #, amount, due date, line items)
- **Payment** — disbursement to satisfy an invoice (bank transfer, check)
- **Payment Terms** — e.g. "Net 30" (pay within 30 days), "2/10 Net 30" (2% discount if paid in 10 days)

Key data model fields for an AP Invoice:
```
vendor_id, company_id, invoice_number, invoice_date, due_date,
amount, currency, status (draft|posted|paid|cancelled),
payment_id (nullable, set when paid), gl_account_id
```

### Accounts Receivable (AR)

AR tracks money **owed by customers**.

Core flow:
```
Sales Order → Delivery → Customer Invoice → Payment Received → Reconciliation
```

Key entities:
- **Customer** — who buys from us (name, CPF/CNPJ, credit limit, payment terms)
- **Sales Invoice** — bill sent to a customer
- **Payment Receipt** — money received from a customer
- **Credit Note** — reduction of a customer's balance (refund, discount)
- **Aging** — how long invoices have been outstanding (30/60/90/120+ days)

### Cash Management

Cash management tracks the actual movement of money through bank accounts.

Key entities:
- **Bank Account** — company's bank accounts (bank, branch, account number, balance)
- **Bank Transaction** — a line from the bank statement
- **Reconciliation** — matching bank transactions to GL entries

---

## Transaction Status Lifecycle

Every financial document should have a status machine:

```
DRAFT → POSTED → PAID/SETTLED
              ↓
            VOID (cannot delete, only void with reverse journal entry)
```

Never hard-delete a financial record. Use `deleted = true` for soft delete, but
voided financial documents must also generate a reversing journal entry.

---

## Multi-Company / Multi-Tenant

relay already has `company_id` on transactions. This is correct. Every financial
entity in an ERP **must always** be scoped to a company:

- Every query must include `WHERE company_id = $1`
- Users belong to one company; they must never see data from another company
- Reports are always per-company (or explicitly cross-company for consolidation)

The `companies` table is the **tenant boundary**. Treat it as inviolable.

---

## Naming Conventions (ERP Standard)

Use these standard names when creating new domains:

| Domain       | Go package  | Table name    |
|--------------|-------------|---------------|
| Transactions | transactions| transactions  |
| Invoices     | invoices    | invoices      |
| Vendors      | vendors     | vendors       |
| Customers    | customers   | customers     |
| Payments     | payments    | payments      |
| GL Accounts  | accounts    | accounts      |
| Journal Entries | journals | journal_entries |
| Bank Accounts| banks       | bank_accounts |
| Fiscal Periods | periods  | fiscal_periods|

---

## Field Naming Standards

Consistent naming across all domains:

| Concept          | Field name         | Type              |
|------------------|--------------------|-------------------|
| Money amount     | `amount`           | NUMERIC(15,2)     |
| Currency code    | `currency`         | CHAR(3) e.g. BRL  |
| Tax ID (company) | `cnpj`             | VARCHAR(14)       |
| Tax ID (person)  | `cpf`              | VARCHAR(11)       |
| Invoice number   | `number`           | VARCHAR(50)       |
| Due date         | `due_date`         | TIMESTAMPTZ       |
| Payment date     | `paid_date`        | TIMESTAMPTZ       |
| Status           | `status`           | INTEGER (enum)    |
| Direction        | `type`             | INTEGER (enum)    |
| Soft delete      | `deleted`          | BOOLEAN DEFAULT false |
| Tenant           | `company_id`       | UUID FK companies |
| Creator          | `creator_id`       | UUID FK users     |

---

## Read More

- `references/financial-flows.md` — Detailed AP/AR/GL flows with SQL schema examples
- `references/erp-roadmap.md` — Suggested feature roadmap from relay's current state to full ERP
