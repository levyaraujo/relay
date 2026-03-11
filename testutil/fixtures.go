package testutil

import (
	"lucrerp/companies"
	"lucrerp/shared"
	"lucrerp/users"
	"time"

	"github.com/google/uuid"
)

func NewCompany() *companies.Company {
	c := &companies.Company{
		Name: "Test Company",
		CNPJ: "12345678000199",
	}
	c.Id = uuid.New()
	c.CreatedAt = time.Now()
	c.UpdatedAt = time.Now()
	return c
}

func NewUser(companyId uuid.UUID) *users.User {
	u := &users.User{
		Name:      "Test User",
		Email:     "test@example.com",
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
