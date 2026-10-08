package main

import (
	"context"
	"database/sql"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"reflect"
	"relay/audit"
	"relay/auth"
	"relay/catalog"
	"relay/finance"
	"relay/infra"
	"relay/organization"
	"relay/party"
	"relay/user"
	"strings"
	"syscall"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
)

var validate *validator.Validate

func main() {
	godotenv.Load(".env")
	validate = validator.New(validator.WithRequiredStructEnabled())

	validate.RegisterTagNameFunc(func(field reflect.StructField) string {
		name := strings.Split(field.Tag.Get("json"), ",")[0]

		if name == "" || name == "-" {
			return field.Name
		}

		return name
	})

	postgresDSN := os.Getenv("POSTGRES_URI")
	db, err := infra.NewDB(postgresDSN)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	mux := setupEndpoints(db)
	dispatcher := audit.NewDispatcher(audit.CreateRepo(db), time.Second)
	ctx, cancel := signalContext()
	defer cancel()
	go dispatcher.Run(ctx)

	slog.Info("Server is running at :8080 🚀")
	log.Fatal(http.ListenAndServe(":8080", middleware(mux)))

}

func signalContext() (context.Context, context.CancelFunc) {
	return signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
}

func middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		slog.Info("API",
			"method", r.Method,
			"path", r.URL.Path,
		)
		next.ServeHTTP(w, r)
	})
}

func setupEndpoints(db *sql.DB) *http.ServeMux {
	userRepo := user.CreateRepo(db)
	orgRepo := organization.CreateRepo(db)
	orgHandler := organization.NewHandler(validate, organization.NewService(orgRepo))
	uHandler := user.NewHandler(validate, user.NewService(userRepo))
	authHandler := auth.NewHandler(validate, auth.NewService(userRepo))
	partyHandler := party.NewHandler(validate, party.NewService(party.CreateRepo(db)))
	itemHandler := catalog.NewHandler(validate, catalog.NewService(catalog.CreateRepo(db)))
	financeHandler := finance.NewHandler(validate, finance.NewService(finance.CreateRepo(db)))
	protected := auth.Middleware(userRepo)

	mux := http.NewServeMux()
	apiMux := http.NewServeMux()

	apiMux.HandleFunc("POST /users", uHandler.CreateUser)
	apiMux.HandleFunc("POST /login", authHandler.Login)
	apiMux.HandleFunc("POST /organizations", orgHandler.CreateOrganization)
	apiMux.Handle("POST /parties", protected(http.HandlerFunc(partyHandler.CreateParty)))
	apiMux.Handle("GET /parties", protected(http.HandlerFunc(partyHandler.ListParties)))
	apiMux.Handle("GET /parties/{id}", protected(http.HandlerFunc(partyHandler.GetParty)))
	apiMux.Handle("POST /items", protected(http.HandlerFunc(itemHandler.CreateItem)))
	apiMux.Handle("GET /items", protected(http.HandlerFunc(itemHandler.ListItems)))
	apiMux.Handle("GET /items/{id}", protected(http.HandlerFunc(itemHandler.GetItem)))
	apiMux.Handle("POST /transactions", protected(http.HandlerFunc(financeHandler.CreateTransaction)))
	apiMux.Handle("GET /transactions", protected(http.HandlerFunc(financeHandler.ListTransactions)))
	apiMux.Handle("GET /transactions/{id}", protected(http.HandlerFunc(financeHandler.GetTransaction)))
	apiMux.Handle("GET /obligations", protected(http.HandlerFunc(financeHandler.ListObligations)))
	apiMux.Handle("GET /obligations/{id}", protected(http.HandlerFunc(financeHandler.GetObligation)))
	apiMux.Handle("POST /obligations/{id}/payments", protected(http.HandlerFunc(financeHandler.CreatePayment)))

	mux.Handle("/api/v1/", http.StripPrefix("/api/v1", apiMux))
	return mux
}
