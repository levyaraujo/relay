package main

import (
	"log"
	"net/http"

	"github.com/levyaraujo/relay/shared"
	"github.com/levyaraujo/relay/transactions"
	"github.com/levyaraujo/relay/vendors"
)

func main() {
	db := shared.Connect()

	transactionsRepo := transactions.NewRepository(db)
	transactionsCtrl := transactions.NewController(transactionsRepo)
	vendorRepo := vendors.NewRepository(db)
	vendorCtrl := vendors.NewController(vendorRepo)

	mux := http.NewServeMux()
	transactions.NewHandler(transactionsCtrl).RegisterRoutes(mux)
	vendors.NewHandler(vendorCtrl).RegisterRoutes(mux)

	log.Fatal(http.ListenAndServe(":8080", mux))
}
