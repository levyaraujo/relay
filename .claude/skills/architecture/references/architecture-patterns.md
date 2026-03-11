# Architecture Deep Dives

## Multi-Tenancy: Row-Level vs Schema-Level vs DB-per-Tenant

relay uses **row-level isolation** (shared schema, `company_id` column). This
is the right choice at this stage. Here's when to reconsider:

| Model              | When to use                                      | relay relevance         |
|--------------------|--------------------------------------------------|-------------------------|
| Row-level (current)| < 10,000 tenants, cost-sensitive, fast growth    | ✅ Use now              |
| Schema-per-tenant  | Compliance needs, large tenants, 100s–1000s      | Future premium tier     |
| DB-per-tenant      | Financial/healthcare regulated, < 500 tenants    | Only if enterprise deals|

**Stick with row-level** until there's a specific enterprise customer requiring
schema or DB isolation for compliance reasons (GDPR data residency, etc.).

### PostgreSQL Row-Level Security (optional future hardening)

If you want the DB itself to enforce isolation (not just the application):

```sql
-- Enable RLS on the table
ALTER TABLE transactions ENABLE ROW LEVEL SECURITY;

-- Create a policy: rows are only visible if company_id matches current setting
CREATE POLICY company_isolation ON transactions
    USING (company_id = current_setting('app.current_company_id')::uuid);

-- In Go, set the setting at the start of each request's DB transaction:
_, err = tx.Exec("SET LOCAL app.current_company_id = $1", companyID.String())
```

This makes it physically impossible to read another tenant's data even if
there's a bug in the application layer. Useful when the codebase grows large.

---

## Context-Based Tenant Propagation

As relay adds authentication, the `company_id` must flow from JWT → context → repo.
Never accept it from request body.

```go
// shared/context.go
type contextKey string

const (
    CompanyIDKey contextKey = "company_id"
    UserIDKey    contextKey = "user_id"
)

func CompanyIDFromContext(ctx context.Context) (uuid.UUID, error) {
    v := ctx.Value(CompanyIDKey)
    if v == nil {
        return uuid.Nil, errors.New("no company_id in context")
    }
    return v.(uuid.UUID), nil
}
```

```go
// shared/middleware.go
func AuthMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        token := r.Header.Get("Authorization")
        claims, err := parseJWT(token)
        if err != nil {
            http.Error(w, "unauthorized", http.StatusUnauthorized)
            return
        }
        ctx := r.Context()
        ctx = context.WithValue(ctx, CompanyIDKey, claims.CompanyID)
        ctx = context.WithValue(ctx, UserIDKey, claims.UserID)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

---

## Audit Log Implementation

### What to Log

Log every mutation to financial data:
- CREATE: log new values
- UPDATE: log old and new values (fields that changed)
- Status change: log old and new status
- Soft delete: log as action="delete"

### Go Implementation Pattern

```go
// shared/audit.go
type AuditAction string

const (
    AuditCreate       AuditAction = "create"
    AuditUpdate       AuditAction = "update"
    AuditDelete       AuditAction = "delete"
    AuditStatusChange AuditAction = "status_change"
)

type AuditLog struct {
    ID         uuid.UUID   `db:"id"`
    CompanyID  uuid.UUID   `db:"company_id"`
    TableName  string      `db:"table_name"`
    RecordID   uuid.UUID   `db:"record_id"`
    Action     AuditAction `db:"action"`
    ActorID    uuid.UUID   `db:"actor_id"`
    OldValues  []byte      `db:"old_values"` // JSON
    NewValues  []byte      `db:"new_values"` // JSON
    OccurredAt time.Time   `db:"occurred_at"`
}

type AuditRepository interface {
    Log(ctx context.Context, entry *AuditLog) error
}
```

```go
// In the controller, after a successful mutation:
func (c *InvoiceController) PostInvoice(ctx context.Context, inv *Invoice) (*Invoice, error) {
    old := *inv
    inv.Status = InvoiceStatusPosted
    inv.UpdatedAt = time.Now()

    updated, err := c.repo.Update(inv)
    if err != nil {
        return nil, err
    }

    userID, _ := shared.UserIDFromContext(ctx)
    companyID, _ := shared.CompanyIDFromContext(ctx)

    oldJSON, _ := json.Marshal(old)
    newJSON, _ := json.Marshal(updated)

    _ = c.auditRepo.Log(ctx, &shared.AuditLog{
        ID:        uuid.New(),
        CompanyID: companyID,
        TableName: "invoices",
        RecordID:  inv.Id,
        Action:    shared.AuditStatusChange,
        ActorID:   userID,
        OldValues: oldJSON,
        NewValues: newJSON,
        OccurredAt: time.Now(),
    })

    return updated, nil
}
```

---

## Idempotency for Financial Operations

Financial operations must be **idempotent** — submitting the same payment twice
must not create two payments.

Approach: use a client-supplied `idempotency_key` (UUID):

```sql
ALTER TABLE payments ADD COLUMN idempotency_key UUID UNIQUE;
```

```go
func (c *PaymentController) CreatePayment(p *Payment) (*Payment, error) {
    // Try to fetch existing payment with this idempotency key
    existing, err := c.repo.GetByIdempotencyKey(p.IdempotencyKey)
    if err == nil {
        return existing, nil // already processed — return it
    }
    if !errors.Is(err, sql.ErrNoRows) {
        return nil, err
    }
    // Not found — process normally
    return c.repo.Create(p)
}
```

---

## Database Transaction Wrapping

When a business operation must update multiple tables atomically (e.g. posting
an invoice: update invoice status + insert journal entry), use a DB transaction:

```go
// repo method that wraps multiple operations
func (r *repository) PostInvoice(inv *Invoice, entry *JournalEntry) error {
    tx, err := r.db.Beginx()
    if err != nil {
        return fmt.Errorf("begin tx: %w", err)
    }
    defer tx.Rollback()

    if _, err := tx.NamedExec(`UPDATE invoices SET status = :status, updated_at = :updated_at WHERE id = :id`, inv); err != nil {
        return fmt.Errorf("update invoice: %w", err)
    }

    if _, err := tx.NamedExec(`INSERT INTO journal_entries (...) VALUES (...)`, entry); err != nil {
        return fmt.Errorf("insert journal: %w", err)
    }

    return tx.Commit()
}
```

---

## Background Jobs for ERP

Some ERP operations should be async, not blocking HTTP responses:
- Sending invoice emails to customers/vendors
- Generating PDF invoices
- Calculating account balances for reporting
- Importing bank statement files
- Running month-end close jobs

When implementing these, add a `jobs` table and a worker:

```sql
CREATE TABLE jobs (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id  UUID NOT NULL,
    type        TEXT NOT NULL,          -- 'send_invoice_email', 'bank_import'
    payload     JSONB NOT NULL,
    status      TEXT NOT NULL DEFAULT 'pending',
    attempts    INTEGER NOT NULL DEFAULT 0,
    scheduled_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    processed_at TIMESTAMPTZ,
    error       TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

Consider libraries: [River](https://riverqueue.com) (Go, Postgres-native) or
[Asynq](https://github.com/hibiken/asynq) (Redis-based).
