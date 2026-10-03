package user

type Role string

const (
	OWNER Role = "OWNER"
	ADMIN Role = "ADMIN"
)

var roleName = map[Role]string{
	OWNER: "OWNER",
	ADMIN: "ADMIN",
}

func (r Role) String() string {
	return roleName[r]
}
