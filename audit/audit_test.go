package audit

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"
	"uuid"

	"github.com/DATA-DOG/go-sqlmock"
)

type recordingRepository struct {
	pending    []Event
	audits     []Event
	published  []uuid.UUID
	failed     []uuid.UUID
	auditError error
}

func (r *recordingRepository) PendingEvents(_ context.Context, _ int) ([]Event, error) {
	return append([]Event(nil), r.pending...), nil
}

func (r *recordingRepository) RecordAudit(_ context.Context, event Event) error {
	if r.auditError != nil {
		return r.auditError
	}
	r.audits = append(r.audits, event)
	return nil
}

func (r *recordingRepository) MarkPublished(_ context.Context, id uuid.UUID) error {
	r.published = append(r.published, id)
	r.pending = r.pending[1:]
	return nil
}

func (r *recordingRepository) MarkFailed(_ context.Context, id uuid.UUID, _ error) error {
	r.failed = append(r.failed, id)
	return nil
}

func TestDispatcherRecordsAuditBeforeMarkingEventPublished(t *testing.T) {
	event := Event{ID: uuid.New(), EntityType: "PARTY", EntityID: uuid.New(), Action: "CREATE", OccurredAt: time.Now()}
	repo := &recordingRepository{pending: []Event{event}}
	dispatcher := Dispatcher{repo: repo, batchSize: 10}

	if err := dispatcher.ProcessOnce(context.Background()); err != nil {
		t.Fatalf("ProcessOnce() error = %v", err)
	}

	if len(repo.audits) != 1 || repo.audits[0].ID != event.ID {
		t.Fatalf("audits = %v, want one audit for %s", repo.audits, event.ID)
	}
	if len(repo.published) != 1 || repo.published[0] != event.ID {
		t.Fatalf("published = %v, want one publication for %s", repo.published, event.ID)
	}
}

func TestRepoEnqueueWritesOutboxEvent(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	event := Event{
		ID:             uuid.New(),
		OrganizationID: uuid.New(),
		EventType:      "PARTY_CREATED",
		EntityType:     "PARTY",
		EntityID:       uuid.New(),
		Action:         "CREATE",
		Payload:        []byte(`{"name":"João"}`),
		OccurredAt:     time.Date(2026, 10, 7, 12, 0, 0, 0, time.UTC),
	}
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO outbox_events")).
		WithArgs(event.ID, event.OrganizationID, event.ActorUserID, event.EventType, event.EntityType, event.EntityID, event.Action, event.Payload, event.OccurredAt).
		WillReturnResult(sqlmock.NewResult(1, 1))

	var exec *sql.DB = db
	if err := (Repo{}).Enqueue(exec, event); err != nil {
		t.Fatalf("Enqueue() error = %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("mock expectations: %v", err)
	}
}

func TestDispatcherLeavesEventUnpublishedWhenAuditFails(t *testing.T) {
	event := Event{ID: uuid.New(), EntityType: "PAYMENT", EntityID: uuid.New(), Action: "CREATE", OccurredAt: time.Now()}
	repo := &recordingRepository{pending: []Event{event}, auditError: errors.New("audit unavailable")}
	dispatcher := Dispatcher{repo: repo, batchSize: 10}

	if err := dispatcher.ProcessOnce(context.Background()); err == nil {
		t.Fatal("ProcessOnce() error = nil, want audit failure")
	}
	if len(repo.published) != 0 {
		t.Fatalf("published = %v, want no published events", repo.published)
	}
	if len(repo.failed) != 1 || repo.failed[0] != event.ID {
		t.Fatalf("failed = %v, want one failure for %s", repo.failed, event.ID)
	}
}
