package main

import (
	"database/sql"
	"log"
	"log/slog"
	"net/http"
	"os"
	"reflect"
	"relay/auth"
	"relay/infra"
	"relay/organization"
	"relay/user"
	"strings"

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

	slog.Info("Server is running at :8080 🚀")
	log.Fatal(http.ListenAndServe(":8080", middleware(mux)))

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

	mux := http.NewServeMux()
	mux.HandleFunc("POST /users", uHandler.CreateUser)
	mux.HandleFunc("POST /login", authHandler.Login)
	mux.HandleFunc("POST /organizations", orgHandler.CreateOrganization)
	return mux
}
