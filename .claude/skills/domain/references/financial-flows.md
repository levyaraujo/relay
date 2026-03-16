# Financial Flows Reference

## AP Flow — Full Detail

### 1. Vendor Invoice Entry

When a vendor bill arrives:

```sql
-- vendors table
CREATE TABLE vendors (
    id          UUID PRIMARY KEY,
    company_id  UUID NOT NULL REFERENCES companies(id),
    name        VARCHAR(255) NOT NULL,
    cnpj        VARCHAR(14),
    email       VARCHAR(255),
    payment_terms INTEGER NOT NULL DEFAULT 30, -- days
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted     BOOLEAN NOT NULL DEFAULT false
);

-- invoices table (AP invoices)
CREATE TABLE invoices (
    id             UUID PRIMARY KEY,
    company_id     UUID NOT NULL REFERENCES companies(id),
    vendor_id      UUID NOT NULL REFERENCES vendors(id),
    number         VARCHAR(50) NOT NULL,
    invoice_date   TIMESTAMPTZ NOT NULL,
    due_date       TIMESTAMPTZ NOT NULL,
    amount         NUMERIC(15,2) NOT NULL,
    currency       CHAR(3) NOT NULL DEFAULT 'BRL',
    status         INTEGER NOT NULL DEFAULT 0, -- 0=draft, 1=posted, 2=paid, 3=void
    description    TEXT NOT NULL DEFAULT '',
    creator_id     UUID NOT NULL REFERENCES users(id),
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted        BOOLEAN NOT NULL DEFAULT false
);
```

### 2. Payment Processing

```sql
-- payments table
CREATE TABLE payments (
    id              UUID PRIMARY KEY,
    company_id      UUID NOT NULL REFERENCES companies(id),
    invoice_id      UUID NOT NULL REFERENCES invoices(id),
    amount          NUMERIC(15,2) NOT NULL,
    payment_date    TIMESTAMPTZ NOT NULL,
    bank_account_id UUID NOT NULL REFERENCES bank_accounts(id),
    reference       VARCHAR(100),     -- bank transfer ID, check number
    status          INTEGER NOT NULL DEFAULT 0,
    creator_id      UUID NOT NULL REFERENCES users(id),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted         BOOLEAN NOT NULL DEFAULT false
);
```

### 3. Status Enum Values (standard across domains)

```go
type InvoiceStatus int

const (
    InvoiceStatusDraft   InvoiceStatus = iota // 0 — not yet posted to GL
    InvoiceStatusPosted                       // 1 — posted to GL, awaiting payment
    InvoiceStatusPaid                         // 2 — fully paid
    InvoiceStatusVoid                         // 3 — cancelled with reversing entry
)
```

---

## AR Flow — Full Detail

Mirrored from AP but from the customer side:

```
Customer → Sales Invoice → Payment Received → Mark as Paid → GL Posting
```

Key difference: AR invoices increase revenue and create an asset (money owed to us).
AP invoices increase expenses and create a liability (money we owe).

```sql
CREATE TABLE customers (
    id           UUID PRIMARY KEY,
    company_id   UUID NOT NULL REFERENCES companies(id),
    name         VARCHAR(255) NOT NULL,
    cnpj         VARCHAR(14),
    cpf          VARCHAR(11),
    email        VARCHAR(255),
    credit_limit NUMERIC(15,2) NOT NULL DEFAULT 0,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted      BOOLEAN NOT NULL DEFAULT false
);
```

---

## Chart of Accounts — Standard Structure

A chart of accounts groups all financial accounts by type. Standard numbering:

| Range     | Type         | Examples                              |
|-----------|--------------|---------------------------------------|
| 1000–1999 | Assets       | Cash, Bank, AR, Inventory             |
| 2000–2999 | Liabilities  | AP, Loans, Tax Payable                |
| 3000–3999 | Equity       | Share Capital, Retained Earnings      |
| 4000–4999 | Revenue      | Sales Revenue, Service Revenue        |
| 5000–5999 | Expenses     | COGS, Salaries, Rent, Utilities       |

```sql
CREATE TABLE accounts (
    id         UUID PRIMARY KEY,
    company_id UUID NOT NULL REFERENCES companies(id),
    code       VARCHAR(20) NOT NULL,         e.g. "1100"  -- 
    name       VARCHAR(255) NOT NULL,          -- e.g. "Cash"
    type       INTEGER NOT NULL,               -- asset/liability/equity/revenue/expense
    parent_id  UUID REFERENCES accounts(id),  -- for hierarchical chart of accounts
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted    BOOLEAN NOT NULL DEFAULT false,
    UNIQUE(company_id, code)
);
```

---

## Double-Entry Journal

Every financial event generates a balanced journal entry (debits = credits):

**Example: Posting a vendor invoice for R$ 1,000**
```
DEBIT  5100 – Office Supplies Expense    1,000.00
CREDIT 2000 – Accounts Payable           1,000.00
```

**Example: Paying that vendor invoice**
```
DEBIT  2000 – Accounts Payable           1,000.00
CREDIT 1100 – Bank Account               1,000.00
```

```sql
CREATE TABLE journal_entries (
    id          UUID PRIMARY KEY,
    company_id  UUID NOT NULL REFERENCES companies(id),
    date        TIMESTAMPTZ NOT NULL,
    description TEXT NOT NULL,
    reference   VARCHAR(100),       -- invoice ID, payment ID, etc.
    creator_id  UUID NOT NULL REFERENCES users(id),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted     BOOLEAN NOT NULL DEFAULT false
);

CREATE TABLE journal_lines (
    id               UUID PRIMARY KEY,
    journal_entry_id UUID NOT NULL REFERENCES journal_entries(id),
    account_id       UUID NOT NULL REFERENCES accounts(id),
    type             INTEGER NOT NULL, -- 1=debit, 2=credit
    amount           NUMERIC(15,2) NOT NULL CHECK (amount > 0),
    description      TEXT
);
-- Constraint: sum of debits must equal sum of credits per journal_entry_id
```

---

## Aging Report Logic

AP aging: how long have we owed vendors money?
AR aging: how long have customers owed us money?

```sql
-- AR Aging buckets
SELECT
    i.id,
    c.name AS customer,
    i.amount,
    i.due_date,
    CURRENT_DATE - i.due_date::date AS days_overdue,
    CASE
        WHEN CURRENT_DATE <= i.due_date::date THEN 'current'
        WHEN CURRENT_DATE - i.due_date::date <= 30 THEN '1-30'
        WHEN CURRENT_DATE - i.due_date::date <= 60 THEN '31-60'
        WHEN CURRENT_DATE - i.due_date::date <= 90 THEN '61-90'
        ELSE '90+'
    END AS aging_bucket
FROM invoices i
JOIN customers c ON c.id = i.customer_id
WHERE i.company_id = $1
  AND i.status = 1  -- posted, not paid
  AND i.deleted = false;
```

---

## relay's transactions Table — Bridging to Full ERP

relay's current `transactions` table is a simplified ledger entry. It can evolve:

**Current**: a single debit/credit record with amount, description, type, origin.

**Path to GL**: add `account_id` (FK to chart of accounts) and `fiscal_period_id`.

**Path to AP/AR**: add `invoice_id` FK — a transaction becomes the payment posting for an invoice.

This means transactions is the **payment/posting layer**, while invoices is the
**document layer** sitting above it. This is the standard ERP separation:
- Invoices = commercial documents
- Transactions/journal lines = accounting postings
