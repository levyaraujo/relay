package user

import "uuid"

type User struct {
	ID             uuid.UUID  `json:"id"`
	Name           string     `json:"name"`
	Document       string     `json:"document"`
	Phone          string     `json:"phone"`
	Email          string     `json:"email"`
	Password       string     `json:"password"`
	OrganizationID *uuid.UUID `json:"organization_id"`
	Role           Role       `json:"role"`
}
