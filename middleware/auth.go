package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/levyaraujo/relay/shared"
	"github.com/levyaraujo/relay/users"
)

func Auth(userRepo users.UserRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if header == "" {
				shared.JSON(w, http.StatusUnauthorized, map[string]string{"error": "missing authorization header"})
				return
			}

			parts := strings.SplitN(header, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				shared.JSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid authorization header"})
				return
			}

			secret := shared.LoadConfig().JWT_SECRET
			token, err := jwt.Parse(parts[1], func(t *jwt.Token) (any, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				return []byte(secret), nil
			})

			if err != nil || !token.Valid {
				shared.JSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid or expired token"})
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				shared.JSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid token claims"})
				return
			}

			sub, _ := claims.GetSubject()
			id, err := uuid.Parse(sub)
			if err != nil {
				shared.JSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid token subject"})
				return
			}

			user, err := userRepo.GetByID(id)
			if err != nil {
				shared.JSON(w, http.StatusUnauthorized, map[string]string{"error": "user not found"})
				return
			}

			ctx := context.WithValue(r.Context(), shared.UserID, id)
			ctx = context.WithValue(ctx, shared.CompanyID, user.CompanyId)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
