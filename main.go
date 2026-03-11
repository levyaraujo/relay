package main

import (
	"log"
	"lucrerp/companies"
	"lucrerp/shared"
	"lucrerp/transactions"
	"lucrerp/users"
	"net/http"
)

func main() {
	db := shared.Connect()

	companies.Migrate()
	users.Migrate()
	transactions.Migrate()

	transactionsRepo := transactions.NewRepository(db)

	mux := http.NewServeMux()
	transactions.NewHandler(transactionsRepo).RegisterRoutes(mux)

	log.Fatal(http.ListenAndServe(":8080", mux))
}
