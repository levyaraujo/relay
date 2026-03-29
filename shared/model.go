package shared

import (
	"time"

	"github.com/google/uuid"
)

type Model struct {
	Id        uuid.UUID `db:"id" json:"id"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"-"`
	Deleted   bool      `db:"deleted" json:"-"`
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
