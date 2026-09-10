package account

import "uuid"

type Account struct {
	ID      uuid.UUID
	Name    string
	Website string
}
