package main

import (
	"log"
	"net/http"

	"github.com/levyaraujo/relay/accounts"
	"github.com/levyaraujo/relay/companies"
	"github.com/levyaraujo/relay/shared"
	"github.com/levyaraujo/relay/transactions"
	"github.com/levyaraujo/relay/vendors"
)

func main() {
	db := shared.Connect()

	companyRepo := companies.NewRepository(db)
	companyCtrl := companies.NewController(companyRepo)
	transactionsRepo := transactions.NewRepository(db)
	transactionsCtrl := transactions.NewController(transactionsRepo)
	vendorRepo := vendors.NewRepository(db)
	vendorCtrl := vendors.NewController(vendorRepo)

	mux := http.NewServeMux()
	accRepo := accounts.NewRepository(db)
	companies.NewHandler(companyCtrl, accRepo).RegisterRoutes(mux)
	transactions.NewHandler(transactionsCtrl).RegisterRoutes(mux)
	vendors.NewHandler(vendorCtrl).RegisterRoutes(mux)

	log.Printf("🚀 Listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
