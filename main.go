package main

import (
	"github.com/levyaraujo/relay/shared"
	"github.com/levyaraujo/relay/transactions"
	"log"
	"net/http"
)

func main() {
	db := shared.Connect()
	transactionsRepo := transactions.NewRepository(db)

	mux := http.NewServeMux()
	transactions.NewHandler(transactionsRepo).RegisterRoutes(mux)

	log.Fatal(http.ListenAndServe(":8080", mux))
}
