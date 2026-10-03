package organization

import "uuid"

type OrgType int

const (
	PRODUCTS OrgType = iota
	SERVICES
)

var typeName = map[OrgType]string{
	PRODUCTS: "PRODUCTS",
	SERVICES: "SERVICES",
}

func (r OrgType) String() string {
	return typeName[r]
}

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
