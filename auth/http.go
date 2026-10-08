package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"relay/shared"
	"relay/user"
	"strings"
	"uuid"

	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
)

type LoginPayload struct {
	Email    string `validate:"required,email" json:"email"`
	Password string `validate:"required" json:"password"`
}

type LoginErrorResponse struct {
	Message string `json:"message"`
}

type JWTResponse struct {
	AccessToken string `json:"accessToken"`
}

type userLookup interface {
	FindByID(uuid.UUID) (user.User, error)
}

type organizationIDContextKey struct{}

func OrganizationIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	organizationID, ok := ctx.Value(organizationIDContextKey{}).(uuid.UUID)
	return organizationID, ok
}

type Handler struct {
	v *validator.Validate
	s *Service
}

func NewHandler(v *validator.Validate, s *Service) *Handler {
	return &Handler{
		v,
		s,
	}
}

func (h Handler) Login(w http.ResponseWriter, r *http.Request) {
	var l LoginPayload

	err := json.NewDecoder(r.Body).Decode(&l)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = h.v.Struct(l)

	if err != nil {
		res := LoginErrorResponse{Message: "Please, provide valid email and password"}

		shared.JSONResponse(w, http.StatusBadRequest, res)
		return
	}

	t, err := h.s.Login(l)

	if err != nil {
		res := LoginErrorResponse{Message: "Something went wrong. Please try again later."}
		shared.JSONResponse(w, http.StatusInternalServerError, res)
		return
	}

	res := JWTResponse{AccessToken: t}

	shared.JSONResponse(w, http.StatusOK, res)
}

func Middleware(lookup userLookup) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if !strings.HasPrefix(authHeader, "Bearer ") {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			claims := &Claims{}
			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			secret := os.Getenv("SECRET_KEY")
			token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (any, error) {
				return []byte(secret), nil
			})
			if err != nil || !token.Valid {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			userID, err := uuid.Parse(claims.Sub)
			if err != nil {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			currentUser, err := lookup.FindByID(userID)
			if err != nil || currentUser.OrganizationID == nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), organizationIDContextKey{}, *currentUser.OrganizationID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

type Claims struct {
	Sub string `json:"sub"`
	jwt.RegisteredClaims
}

func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	godotenv.Load("../.env")
	SECRET := os.Getenv("SECRET_KEY")

	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (any, error) {
			return SECRET, nil
		})
		if err != nil || !token.Valid {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}
		// Add username to context for downstream use
		ctx := context.WithValue(r.Context(), "userId", claims.Sub)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}
