package user

import (
	"database/sql"
)

type Repo struct {
	db *sql.DB
}

func (r Repo) Create(u UserPayload) *sql.Row {
	insert := `INSERT INTO users (name, document, phone, email, password) VALUES ($1, $2, $3, $4, $5) RETURNING email`

	return r.db.QueryRow(insert, u.Name, u.Document, u.Phone, u.Email, u.Password)
}

func (r Repo) FindByEmail(email string) (User, error) {
	var u User
	query := `SELECT * FROM users WHERE email = $1`

	err := r.db.QueryRow(query, email).Scan(
		&u.ID,
		&u.Name,
		&u.Document,
		&u.Phone,
		&u.Email,
		&u.Password,
	)

	return u, err
}

func CreateRepo(db *sql.DB) Repo {
	return Repo{db}
}
