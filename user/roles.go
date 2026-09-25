package user

type Role int

const (
	OWNER Role = iota
	ADMIN
)

var roleName = map[Role]string{
	OWNER: "OWNER",
	ADMIN: "ADMIN",
}

func (r Role) String() string {
	return roleName[r]
}
