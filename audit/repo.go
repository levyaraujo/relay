package audit

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"
	"uuid"
)

type repository interface {
	PendingEvents(context.Context, int) ([]Event, error)
	RecordAudit(context.Context, Event) error
	MarkPublished(context.Context, uuid.UUID) error
	MarkFailed(context.Context, uuid.UUID, error) error
}

type SQLExecutor interface {
	Exec(query string, args ...any) (sql.Result, error)
}

type Repo struct {
	db *sql.DB
}

func (r Repo) Enqueue(exec SQLExecutor, event Event) error {
	if event.ID == uuid.Nil() {
		event.ID = uuid.New()
	}
	if event.OccurredAt.IsZero() {
		event.OccurredAt = time.Now()
	}
	if len(event.Payload) == 0 {
		event.Payload = json.RawMessage(`{}`)
	}

	_, err := exec.Exec(`
		INSERT INTO outbox_events
			(id, organization_id, actor_user_id, event_type, entity_type, entity_id, action, payload, occurred_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		event.ID,
		event.OrganizationID,
		event.ActorUserID,
		event.EventType,
		event.EntityType,
		event.EntityID,
		event.Action,
		event.Payload,
		event.OccurredAt,
	)
	return err
}

func (r Repo) PendingEvents(ctx context.Context, limit int) ([]Event, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, organization_id, actor_user_id, event_type, entity_type, entity_id, action, payload, occurred_at
		FROM outbox_events
		WHERE published_at IS NULL
		ORDER BY occurred_at, id
		LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []Event
	for rows.Next() {
		var event Event
		var actorUserID sql.NullString
		if err := rows.Scan(
			&event.ID,
			&event.OrganizationID,
			&actorUserID,
			&event.EventType,
			&event.EntityType,
			&event.EntityID,
			&event.Action,
			&event.Payload,
			&event.OccurredAt,
		); err != nil {
			return nil, err
		}
		if actorUserID.Valid {
			id, err := uuid.Parse(actorUserID.String)
			if err != nil {
				return nil, err
			}
			event.ActorUserID = &id
		}
		events = append(events, event)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return events, nil
}

func (r Repo) RecordAudit(ctx context.Context, event Event) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO audit_logs
			(event_id, organization_id, actor_user_id, entity_type, entity_id, action, changes)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (event_id) DO NOTHING`,
		event.ID,
		event.OrganizationID,
		event.ActorUserID,
		event.EntityType,
		event.EntityID,
		event.Action,
		event.Payload,
	)
	return err
}

func (r Repo) MarkPublished(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE outbox_events
		SET published_at = now(), attempts = attempts + 1, last_error = NULL
		WHERE id = $1`, id)
	return err
}

func (r Repo) MarkFailed(ctx context.Context, id uuid.UUID, processingErr error) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE outbox_events
		SET attempts = attempts + 1, last_error = $2
		WHERE id = $1`, id, processingErr.Error())
	return err
}

func CreateRepo(db *sql.DB) Repo {
	return Repo{db: db}
}
