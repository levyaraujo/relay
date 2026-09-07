package user

import "uuid"

type User struct {
	ID       uuid.UUID
	Name     string
	Document string
	Phone    string
	Email    string
	Password string
}
