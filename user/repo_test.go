package user

import (
	"regexp"
	"testing"
	"uuid"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestFindByIDLoadsUserOrganization(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	userID := uuid.New()
	organizationID := uuid.New()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM users WHERE id = $1")).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "name", "document", "phone", "email", "password", "organization_id", "role",
		}).AddRow(userID, "Owner", "52998224725", "5511999999999", "owner@example.com", "hash", organizationID, "OWNER"))

	got, err := (Repo{db: db}).FindByID(userID)
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}
	if got.ID != userID {
		t.Fatalf("ID = %s, want %s", got.ID, userID)
	}
	if got.OrganizationID == nil || *got.OrganizationID != organizationID {
		t.Fatalf("OrganizationID = %v, want %s", got.OrganizationID, organizationID)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("mock expectations: %v", err)
	}
}
