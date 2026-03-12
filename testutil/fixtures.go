package testutil

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/levyaraujo/relay/companies"
	"github.com/levyaraujo/relay/shared"
	"github.com/levyaraujo/relay/users"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

func NewCompany() *companies.Company {
	c := &companies.Company{
		Name: "Test Company",
		CNPJ: fmt.Sprintf("%014d", rand.Int63n(99999999999999)),
	}
	c.Id = uuid.New()
	c.CreatedAt = time.Now()
	c.UpdatedAt = time.Now()
	return c
}

func NewUser(companyId uuid.UUID) *users.User {
	u := &users.User{
		Name:      "Test User",
		Email:     uuid.New().String() + "@test.com",
		Password:  "hashed_password",
		CompanyId: companyId,
	}
	u.Id = uuid.New()
	u.CreatedAt = time.Now()
	u.UpdatedAt = time.Now()
	return u
}

func SeedCompanyAndUser(companyRepo companies.CompanyRepository, userRepo users.UserRepository) (*companies.Company, *users.User) {
	c, err := companyRepo.Create(NewCompany())
	if err != nil {
		panic("failed to seed company: " + err.Error())
	}

	u, err := userRepo.Create(NewUser(c.Id))
	if err != nil {
		panic("failed to seed user: " + err.Error())
	}

	return c, u
}

func SetupDB() *shared.Settings {
	return shared.LoadConfig()
}

// CleanTables deletes all rows from the given tables in order.
// Pass table names in reverse FK dependency order (children first).
// t is optional — pass nil when calling from TestMain.
func CleanTables(t testing.TB, db *sqlx.DB, tables ...string) {
	if t != nil {
		t.Helper()
	}
	for _, table := range tables {
		db.MustExec("DELETE FROM " + table)
	}
}
