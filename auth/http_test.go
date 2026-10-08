package auth

import (
	"net/http"
	"net/http/httptest"
	"os"
	"relay/user"
	"testing"
	"uuid"

	"github.com/golang-jwt/jwt/v5"
)

type userLookupStub struct {
	user user.User
	err  error
}

func (s userLookupStub) FindByID(uuid.UUID) (user.User, error) {
	return s.user, s.err
}

func TestMiddlewareAddsOrganizationIDToRequestContext(t *testing.T) {
	previousSecret := os.Getenv("SECRET_KEY")
	t.Cleanup(func() { _ = os.Setenv("SECRET_KEY", previousSecret) })
	_ = os.Setenv("SECRET_KEY", "test-secret")

	organizationID := uuid.New()
	userID := uuid.New()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"sub": userID.String()})
	tokenString, err := token.SignedString([]byte("test-secret"))
	if err != nil {
		t.Fatalf("SignedString() error = %v", err)
	}

	request := httptest.NewRequest(http.MethodGet, "/parties", nil)
	request.Header.Set("Authorization", "Bearer "+tokenString)
	response := httptest.NewRecorder()
	var gotOrganizationID uuid.UUID

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var ok bool
		gotOrganizationID, ok = OrganizationIDFromContext(r.Context())
		if !ok {
			t.Error("organization ID missing from request context")
		}
		w.WriteHeader(http.StatusNoContent)
	})

	handler := Middleware(userLookupStub{user: user.User{ID: userID, OrganizationID: &organizationID}})(next)
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusNoContent)
	}
	if gotOrganizationID != organizationID {
		t.Fatalf("organization ID = %s, want %s", gotOrganizationID, organizationID)
	}
}

func TestMiddlewareRejectsUserWithoutOrganization(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/parties", nil)
	request.Header.Set("Authorization", "Bearer invalid")
	response := httptest.NewRecorder()

	handler := Middleware(userLookupStub{})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("next handler should not be called")
	}))
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusUnauthorized)
	}
}
