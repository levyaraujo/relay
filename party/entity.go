package party

import "uuid"

type Role string

const (
	CustomerRole Role = "CUSTOMER"
	SupplierRole Role = "SUPPLIER"
)

type Party struct {
	ID             uuid.UUID `json:"id"`
	OrganizationID uuid.UUID `json:"organizationId"`
	Name           string    `json:"name"`
	Email          string    `json:"email"`
	Phone          string    `json:"phone"`
	Document       string    `json:"document"`
	Roles          []Role    `json:"roles"`
}
