package account

import (
	"database/sql"
)

type Repo struct {
	db *sql.DB
}

func (r Repo) Create(a Account) *sql.Row {
	insert := `INSERT INTO account (name, website) VALUES ($1, $2)`
	return r.db.QueryRow(insert, a.Name, a.Website)
}
