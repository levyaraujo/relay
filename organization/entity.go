package organization

import "uuid"

type OrgType int

const (
	PRODUCT OrgType = iota
	SERVICE
)

var roleName = map[OrgType]string{
	PRODUCT: "OWNER",
	SERVICE: "ADMIN",
}

func (r OrgType) String() string {
	return roleName[r]
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
