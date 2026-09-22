package main

import (
	"log"
	"log/slog"
	"net/http"
	"os"
	"relay/infra"
	"relay/user"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
)

var validate *validator.Validate

func main() {
	godotenv.Load(".env")
	validate = validator.New(validator.WithRequiredStructEnabled())

	postgresDSN := os.Getenv("POSTGRES_URI")
	db, err := infra.NewDB(postgresDSN)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	userRepo := user.CreateRepo(db)
	createUserHandler := user.NewHandler(validate, user.NewService(userRepo)).CreateUser

	mux := http.NewServeMux()
	mux.HandleFunc("POST /users", createUserHandler)

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
