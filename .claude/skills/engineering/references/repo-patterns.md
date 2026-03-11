# Repository Patterns Reference

## sqlx NamedQuery + RETURNING

Use this pattern for all INSERT and UPDATE statements that need to return the full row:

```go
func (r *repository) Create(t *Transaction) (*Transaction, error) {
    var created Transaction
    rows, err := r.db.NamedQuery(`
        INSERT INTO transactions (id, company_id, type, amount, ...)
        VALUES (:id, :company_id, :type, :amount, ...)
        RETURNING *
    `, t)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    if rows.Next() {
        if err := rows.StructScan(&created); err != nil {
            return nil, err
        }
    }
    return &created, nil
}
```

**Why not `db.Get` or `db.QueryRowx`?** `NamedQuery` maps struct field tags (`:field_name`) to named parameters automatically, avoiding positional argument bugs.

## Handling NULL / Optional Fields

For fields like `due_date` and `paid_date` that are nullable in the DB, use pointer types:

```go
type Transaction struct {
    shared.Model
    DueDate  *time.Time `db:"due_date"`
    PaidDate *time.Time `db:"paid_date"`
}
```

This allows `nil` to be stored as `NULL` in Postgres. When decoding JSON, a missing field will naturally be `nil`.

## GetByID with soft-delete

Always filter `deleted = false`:

```go
func (r *repository) GetByID(id uuid.UUID) (*Transaction, error) {
    var t Transaction
    err := r.db.Get(&t, `SELECT * FROM transactions WHERE id = $1 AND deleted = false`, id)
    if err != nil {
        return nil, err  // sql.ErrNoRows is surfaced here
    }
    return &t, nil
}
```

To distinguish "not found" from other DB errors in the handler:

```go
import "database/sql"

if errors.Is(err, sql.ErrNoRows) {
    http.Error(w, "not found", http.StatusNotFound)
    return
}
http.Error(w, err.Error(), http.StatusInternalServerError)
```

## Soft Delete

Never hard-delete rows. Implement delete as an update:

```go
func (r *repository) Delete(id uuid.UUID) error {
    _, err := r.db.Exec(
        `UPDATE transactions SET deleted = true, updated_at = now() WHERE id = $1`,
        id,
    )
    return err
}
```

## Listing with Filters

For list endpoints, use positional params to build WHERE clauses:

```go
func (r *repository) ListByCompany(companyID uuid.UUID) ([]*Transaction, error) {
    var ts []*Transaction
    err := r.db.Select(&ts,
        `SELECT * FROM transactions WHERE company_id = $1 AND deleted = false ORDER BY created_at DESC`,
        companyID,
    )
    if err != nil {
        return nil, err
    }
    return ts, nil
}
```

## Updating timestamps

Always update `updated_at` on every UPDATE. Either set it in Go before passing to the repo:

```go
t.UpdatedAt = time.Now()
return c.repo.Update(t)
```

Or in SQL:

```sql
UPDATE transactions SET ..., updated_at = now() WHERE id = :id RETURNING *
```

Prefer the SQL approach so the DB is the authoritative clock.

## Error wrapping

Wrap errors with context when propagating through layers:

```go
import "fmt"

func (r *repository) GetByID(id uuid.UUID) (*Transaction, error) {
    var t Transaction
    err := r.db.Get(&t, `SELECT * FROM transactions WHERE id = $1 AND deleted = false`, id)
    if err != nil {
        return nil, fmt.Errorf("transactions.GetByID %s: %w", id, err)
    }
    return &t, nil
}
```

This preserves the original error for `errors.Is` checks while adding diagnostic context.
