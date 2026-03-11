package users

import (
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type UserRepository interface {
	Create(u *User) (*User, error)
	Update(u *User) (*User, error)
	GetByID(id uuid.UUID) (*User, error)
	GetByEmail(email string) (*User, error)
}

type repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) UserRepository {
	return &repository{db: db}
}

func (r *repository) Create(u *User) (*User, error) {
	var created User
	rows, err := r.db.NamedQuery(`
		INSERT INTO users
			(id, name, email, password, company_id, created_at, updated_at, deleted)
		VALUES
			(:id, :name, :email, :password, :company_id, :created_at, :updated_at, :deleted)
		RETURNING *
	`, u)
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

func (r *repository) Update(u *User) (*User, error) {
	var updated User
	rows, err := r.db.NamedQuery(`
		UPDATE users SET
			name       = :name,
			email      = :email,
			password   = :password,
			updated_at = :updated_at,
			deleted    = :deleted
		WHERE id = :id
		RETURNING *
	`, u)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if rows.Next() {
		if err := rows.StructScan(&updated); err != nil {
			return nil, err
		}
	}
	return &updated, nil
}

func (r *repository) GetByID(id uuid.UUID) (*User, error) {
	var u User
	err := r.db.Get(&u, `SELECT * FROM users WHERE id = $1 AND deleted = false`, id)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *repository) GetByEmail(email string) (*User, error) {
	var u User
	err := r.db.Get(&u, `SELECT * FROM users WHERE email = $1 AND deleted = false`, email)
	if err != nil {
		return nil, err
	}
	return &u, nil
}
