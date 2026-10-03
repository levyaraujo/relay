package organization

import "uuid"

type OrgType string

const (
	PRODUCTS OrgType = "PRODUCTS"
	SERVICES         = "SERVICES"
)

type Organization struct {
	ID          uuid.UUID
	Name        string
	Website     string
	Phone       string
	Email       string
	TaxID       string
	Currency    string
	Type        OrgType
	Description string
}
