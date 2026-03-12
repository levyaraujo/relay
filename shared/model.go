package shared

import (
	"time"

	"github.com/google/uuid"
)

type Model struct {
	Id        uuid.UUID `db:"id"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
	Deleted   bool      `db:"deleted"`
}

// NewModel returns a Model with Id, CreatedAt, and UpdatedAt pre-filled.
func NewModel() Model {
	now := time.Now()
	return Model{
		Id:        uuid.New(),
		CreatedAt: now,
		UpdatedAt: now,
	}
}
