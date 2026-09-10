package account

import (
	"database/sql"
)

type AccountRepo struct {
	db *sql.DB
}

func (r AccountRepo) Create(a *Account) *sql.Row {
	insert := `INSERT INTO account (name, website) VALUES ($1, $2)`
	return r.db.QueryRow(insert, a.Name, a.Website)
}
