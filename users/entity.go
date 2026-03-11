package users

import (
	"github.com/levyaraujo/relay/shared"

	"github.com/google/uuid"
)

type User struct {
	shared.Model
	Name      string    `db:"name"`
	Email     string    `db:"email"`
	Password  string    `db:"password"`
	CompanyId uuid.UUID `db:"company_id"`
}
