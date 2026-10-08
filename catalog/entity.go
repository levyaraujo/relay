package catalog

import "uuid"

type Kind string

const (
	ProductKind Kind = "PRODUCT"
	ServiceKind Kind = "SERVICE"
)

type Item struct {
	ID                uuid.UUID `json:"id"`
	OrganizationID    uuid.UUID `json:"organizationId"`
	Name              string    `json:"name"`
	Description       string    `json:"description"`
	Kind              Kind      `json:"kind"`
	DefaultPriceCents int64     `json:"defaultPriceCents"`
	Currency          string    `json:"currency"`
}
