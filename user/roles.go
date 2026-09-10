package user

import "fmt"

type Role int

const (
	OWNER Role = iota
	ADMIN
	DEVELOPER
)

var roleName = map[Role]string{
	OWNER:     "OWNER",
	ADMIN:     "ADMIN",
	DEVELOPER: "DEVELOPER",
}

func (r Role) String() string {
	return roleName[r]
}

func main() {
	fmt.Println(OWNER)
}
