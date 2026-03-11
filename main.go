package main

import (
	"log"
	"lucrerp/shared"
	"lucrerp/transactions"
	"net/http"
)

func main() {
	db := shared.Connect()
	transactionsRepo := transactions.NewRepository(db)

	mux := http.NewServeMux()
	transactions.NewHandler(transactionsRepo).RegisterRoutes(mux)

	log.Fatal(http.ListenAndServe(":8080", mux))
}
