---
name: architecture
description: >
  ERP system architecture patterns for the relay project. Use this skill when
  making architectural decisions — multi-tenancy, company data isolation, audit
  trails, soft delete, modular domain design, API design, fiscal period enforcement,
  status machines, database schema design for financial data, or scaling a small ERP
  toward production. Triggers on: how should I architect X, is this design correct,
  multi-tenancy approach, should I use company_id or schemas, audit log design,
  transaction isolation, how to enforce company boundaries, how to model status
  lifecycle, how to scale this, what pattern should I use for Y.
---

# ERP Architecture Patterns

Architectural decision reference for **relay**. These patterns are informed by
how modern ERP SaaS products (NetSuite, Odoo, modern fintech) are built.

---

## Tenant Isolation — relay's Approach

relay uses **row-level isolation**: every table has a `company_id` column.
This is the right choice for a small-to-mid-scale ERP SaaS.

### The Golden Rule

> **Every query that touches business data must filter by `company_id`.**

```go
// WRONG — leaks data across companies
r.db.Get(&t, `SELECT * FROM transactions WHERE id = $1`, id)

// CORRECT
r.db.Get(&t, `SELECT * FROM transactions WHERE id = $1 AND company_id = $2 AND deleted = false`, id, companyID)
```

The `company_id` must come from the **authenticated session** (JWT claim, context
value), never from user input. Never trust a `company_id` from a request body.

### Enforcing Tenant Context in Go

Pass company ID via `context.Context`:

```go
// middleware sets company ID from JWT
func TenantMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        companyID := extractCompanyFromJWT(r)
        ctx := context.WithValue(r.Context(), CompanyIDKey, companyID)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

// handler reads it from context
func (h *Handler) getByID(w http.ResponseWriter, r *http.Request) {
    companyID := r.Context().Value(CompanyIDKey).(uuid.UUID)
    id := uuid.MustParse(r.PathValue("id"))
    t, err := h.repo.GetByID(id, companyID)  // always pass both
    // ...
}

// repo always receives and uses companyID
func (r *repository) GetByID(id, companyID uuid.UUID) (*Transaction, error) {
    var t Transaction
    err := r.db.Get(&t,
        `SELECT * FROM transactions WHERE id = $1 AND company_id = $2 AND deleted = false`,
        id, companyID,
    )
    return &t, err
}
```

---

## Soft Delete

relay already uses `deleted = boolean`. Rules:

1. **Never hard delete business data.** Set `deleted = true`.
2. **All SELECT queries must include `AND deleted = false`.**
3. Financial documents (invoices, payments, journal entries) must **also** be voidable — voiding generates a reversing journal entry. Soft delete alone is not enough for financial records.
4. Add `deleted_at TIMESTAMPTZ` for audit purposes if needed.

```go
// Every repo method that lists or fetches must filter:
WHERE company_id = $1 AND deleted = false
```

---

## Status Machines for Financial Documents

Every financial document (invoice, payment, journal entry) must have a status
with explicit allowed transitions. Encode this in the controller layer.

```go
type InvoiceStatus int

const (
    InvoiceStatusDraft  InvoiceStatus = iota // editable
    InvoiceStatusPosted                      // read-only, in GL
    InvoiceStatusPaid                        // settled
    InvoiceStatusVoid                        // cancelled
)

var validTransitions = map[InvoiceStatus][]InvoiceStatus{
    InvoiceStatusDraft:  {InvoiceStatusPosted, InvoiceStatusVoid},
    InvoiceStatusPosted: {InvoiceStatusPaid, InvoiceStatusVoid},
    InvoiceStatusPaid:   {},  // terminal — cannot change a paid invoice
    InvoiceStatusVoid:   {},  // terminal
}

var ErrInvalidTransition = errors.New("invalid status transition")

func (c *InvoiceController) Transition(inv *Invoice, to InvoiceStatus) error {
    allowed := validTransitions[inv.Status]
    for _, s := range allowed {
        if s == to {
            inv.Status = to
            inv.UpdatedAt = time.Now()
            _, err := c.repo.Update(inv)
            return err
        }
    }
    return fmt.Errorf("%w: %s → %s", ErrInvalidTransition, inv.Status, to)
}
```

---

## Audit Trail

Every change to a financial record must be traceable. relay currently captures
`creator_id` and `created_at`. A complete audit trail adds:

```sql
-- Option A: append-only event log (recommended for financial data)
CREATE TABLE audit_log (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id  UUID NOT NULL,
    table_name  TEXT NOT NULL,
    record_id   UUID NOT NULL,
    action      TEXT NOT NULL,         -- 'create', 'update', 'delete', 'status_change'
    actor_id    UUID NOT NULL,         -- user who did it
    old_values  JSONB,
    new_values  JSONB,
    occurred_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_audit_log_record ON audit_log(table_name, record_id);
CREATE INDEX idx_audit_log_company ON audit_log(company_id, occurred_at DESC);
```

Write to `audit_log` in the controller on every mutation. This is append-only —
never update or delete audit log rows.

---

## Modular Domain Architecture

relay's package-per-domain structure scales well. Key rules:

### Domain Boundaries

Each domain package owns its own:
- Database table(s)
- Repository interface
- Business logic (controller)
- HTTP handlers
- Migrations

Domains communicate **only through their public interfaces** — never by
importing another domain's repo struct or querying its table directly.

```
// WRONG — transactions domain reaches into invoices table directly
r.db.Get(&invoice, "SELECT * FROM invoices WHERE id = $1", invoiceID)

// CORRECT — use invoices domain's repository interface
type TransactionController struct {
    repo         TransactionRepository
    invoiceRepo  invoices.InvoiceRepository  // inject the interface
}
```

### When to Create a New Domain

Create a new package when:
- The entity has its own lifecycle (status machine, CRUD)
- It has meaningful business rules that belong together
- It will grow to 3+ operations

Keep in the same package when:
- It's a detail/line item of the parent (e.g. `journal_lines` inside `journals`)
- It has no independent lifecycle

---

## API Design Patterns

relay uses stdlib `net/http`. Conventions to follow as it grows:

### URL Structure

```
POST   /companies/:id/transactions      ← scoped to company
GET    /companies/:id/transactions      ← list with filters
GET    /transactions/:id               ← fetch by ID (company from context)
PUT    /transactions/:id               ← update
DELETE /transactions/:id               ← soft delete (sets deleted=true)
POST   /transactions/:id/void          ← status transition as sub-resource
POST   /transactions/:id/post          ← status transition
```

### Filtering & Pagination

Standard query params for list endpoints:
```
?page=1&per_page=50
?from=2025-01-01&to=2025-12-31    ← date range
?status=1                          ← filter by status
?account_id=<uuid>                 ← filter by GL account
```

### Error Responses

Standardize error JSON:
```json
{
  "error": "amount cannot be negative",
  "code": "INVALID_AMOUNT"
}
```

---

## Financial Data Integrity Rules

These must be enforced at the **database level**, not only in Go:

```sql
-- Amount must always be positive (sign is captured by type/account)
amount NUMERIC(15,2) NOT NULL CHECK (amount > 0)

-- Journal entry must balance (trigger or enforced via application)
-- Total debits = total credits per journal_entry_id

-- Foreign key integrity — always use ON DELETE RESTRICT for financial data
CONSTRAINT fk_company FOREIGN KEY (company_id) REFERENCES companies(id) ON DELETE RESTRICT
CONSTRAINT fk_creator FOREIGN KEY (creator_id) REFERENCES users(id)     ON DELETE RESTRICT
```

Add DB indexes on all FK columns and all commonly filtered columns:
```sql
CREATE INDEX idx_transactions_company_id ON transactions(company_id);
CREATE INDEX idx_transactions_due_date   ON transactions(due_date) WHERE deleted = false;
CREATE INDEX idx_transactions_status     ON transactions(status, company_id) WHERE deleted = false;
```

---

## Scaling Considerations

relay's current architecture (single DB, no framework) scales well to:
- Hundreds of companies
- Millions of transactions per company
- Dozens of concurrent users

When to revisit:
- **Read replicas**: when reporting queries slow down OLTP — connect reports to a read replica
- **Pagination**: add cursor-based pagination before any list endpoint returns > 1000 rows
- **Connection pooling**: add PgBouncer before reaching > 100 concurrent connections
- **Background jobs**: use a job queue (e.g. River, Asynq) for sending emails, generating PDFs, calculating reports asynchronously

---

## Checklist for New Financial Features

Before shipping any financial feature:

- [ ] Every table has `company_id` FK to companies
- [ ] Every query filters `company_id = $1` from context (not from request body)
- [ ] Every query filters `deleted = false`
- [ ] Amount fields use `NUMERIC(15,2)` with `CHECK (amount > 0)`
- [ ] Status lifecycle defined with explicit valid transitions
- [ ] Mutations write to `audit_log` with old/new values
- [ ] FK constraints use `ON DELETE RESTRICT`
- [ ] Indexes created for `company_id` and common filter columns
- [ ] Posted/paid documents cannot be edited (enforced in controller)
- [ ] Void creates a reversing entry, not a hard delete
