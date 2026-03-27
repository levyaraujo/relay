// Command seed creates a default company and user for local development.
package main

import (
	"fmt"
	"log"

	"github.com/levyaraujo/relay/accounts"
	"github.com/levyaraujo/relay/companies"
	"github.com/levyaraujo/relay/shared"
	"github.com/levyaraujo/relay/users"
)

func main() {
	db := shared.Connect()
	defer db.Close()

	companyRepo := companies.NewRepository(db)
	companyCtrl := companies.NewController(companyRepo)
	accRepo := accounts.NewRepository(db)
	userRepo := users.NewRepository(db)
	userCtrl := users.NewController(userRepo)

	co := &companies.Company{
		Name: "Acme Ltda",
		CNPJ: "11222333000181",
	}

	created, err := companyCtrl.Create(co, accRepo)
	if err != nil {
		log.Fatalf("seed company: %v", err)
	}
	fmt.Printf("company created: %s (%s)\n", created.Name, created.Id)

	u := &users.User{
		Model:     shared.NewModel(),
		Name:      "Admin",
		Email:     "admin@acme.com",
		Password:  "changeme123",
		CompanyId: created.Id,
	}

	createdUser, err := userCtrl.Create(u)
	if err != nil {
		log.Fatalf("seed user: %v", err)
	}
	fmt.Printf("user created: %s <%s> (%s)\n", createdUser.Name, createdUser.Email, createdUser.Id)
}
