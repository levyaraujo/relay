package user

import "uuid"

type User struct {
	ID       uuid.UUID `json:"id"`
	Name     string    `json:"name"`
	Document string    `json:"document"`
	Phone    string    `json:"phone"`
	Email    string    `json:"email"`
	Password string    `json:"password"`
}

type AccountUser struct {
	ID        uuid.UUID `json:"id"`
	UserId    uuid.UUID `json:"user_id"`
	AccountId uuid.UUID `json:"account_id"`
	Role      Role      `json:"role"`
}
