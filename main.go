package main

import (
	"log"
	"net/http"

	"github.com/levyaraujo/relay/accounts"
	"github.com/levyaraujo/relay/auth"
	"github.com/levyaraujo/relay/companies"
	"github.com/levyaraujo/relay/middleware"
	"github.com/levyaraujo/relay/shared"
	"github.com/levyaraujo/relay/transactions"
	"github.com/levyaraujo/relay/users"
	"github.com/levyaraujo/relay/vendors"
)

func main() {
	db := shared.Connect()

	userRepo := users.NewRepository(db)
	companyRepo := companies.NewRepository(db)
	accRepo := accounts.NewRepository(db)
	companyCtrl := companies.NewController(companyRepo)
	transactionsRepo := transactions.NewRepository(db)
	transactionsCtrl := transactions.NewController(transactionsRepo)
	vendorRepo := vendors.NewRepository(db)
	vendorCtrl := vendors.NewController(vendorRepo)
	authCtrl := auth.NewController(userRepo)

	public := http.NewServeMux()
	auth.NewHandler(authCtrl).RegisterRoutes(public)

	userCtrl := users.NewController(userRepo)

	protected := http.NewServeMux()
	users.NewHandler(userCtrl).RegisterRoutes(protected)
	companies.NewHandler(companyCtrl, accRepo).RegisterRoutes(protected)
	transactions.NewHandler(transactionsCtrl).RegisterRoutes(protected)
	vendors.NewHandler(vendorCtrl).RegisterRoutes(protected)

	mux := http.NewServeMux()
	mux.Handle("/", public)
	mux.Handle("/api/", middleware.Auth(userRepo)(http.StripPrefix("/api", protected)))

	stack := middleware.Chain(
		middleware.Cors,
		middleware.Logging,
	)

	log.Printf("🚀 Listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", stack(mux)))
}
