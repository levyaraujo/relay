package users

import (
	"github.com/levyaraujo/relay/shared"

	"github.com/google/uuid"
)

type User struct {
	shared.Model
	Name      string    `db:"name" json:"name"`
	Email     string    `db:"email" json:"email"`
	Password  string    `db:"password" json:"-"`
	CompanyId uuid.UUID `db:"company_id" json:"company"`
}
