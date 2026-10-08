package audit

import (
	"context"
	"errors"
	"log/slog"
	"time"
	"uuid"
)

type Dispatcher struct {
	repo      repository
	batchSize int
	interval  time.Duration
}

func NewDispatcher(repo Repo, interval time.Duration) *Dispatcher {
	return &Dispatcher{repo: repo, batchSize: 100, interval: interval}
}

func (d Dispatcher) ProcessOnce(ctx context.Context) error {
	events, err := d.repo.PendingEvents(ctx, d.batchSize)
	if err != nil {
		return err
	}

	for _, event := range events {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := d.repo.RecordAudit(ctx, event); err != nil {
			if markErr := d.repo.MarkFailed(ctx, event.ID, err); markErr != nil {
				return errors.Join(err, markErr)
			}
			return err
		}
		if err := d.repo.MarkPublished(ctx, event.ID); err != nil {
			return err
		}
	}
	return nil
}

func (d Dispatcher) Run(ctx context.Context) {
	if d.interval <= 0 {
		d.interval = time.Second
	}
	ticker := time.NewTicker(d.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := d.ProcessOnce(ctx); err != nil && !errors.Is(err, context.Canceled) {
				slog.Error("audit dispatcher", "err", err)
			}
		}
	}
}

func NewEvent(organizationID uuid.UUID, actorUserID *uuid.UUID, eventType, entityType string, entityID uuid.UUID, action string, payload []byte) Event {
	return Event{
		ID:             uuid.New(),
		OrganizationID: organizationID,
		ActorUserID:    actorUserID,
		EventType:      eventType,
		EntityType:     entityType,
		EntityID:       entityID,
		Action:         action,
		Payload:        payload,
		OccurredAt:     time.Now(),
	}
}
