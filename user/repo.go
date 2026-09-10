package user

import (
	"database/sql"
)

type Repo struct {
	db *sql.DB
}

func (r Repo) Create(u *User) *sql.Row {
	insert := `INSERT INTO users (name, document, phone, email, password) VALUES ($1, $2, $3, $4, $5) RETURNING email`

	return r.db.QueryRow(insert, u.Name, u.Document, u.Phone, u.Email, u.Password)
}
