package audit

import (
	"encoding/json"
	"time"
	"uuid"
)

type Event struct {
	ID             uuid.UUID
	OrganizationID uuid.UUID
	ActorUserID    *uuid.UUID
	EventType      string
	EntityType     string
	EntityID       uuid.UUID
	Action         string
	Payload        json.RawMessage
	OccurredAt     time.Time
}

type AuditLog struct {
	ID             uuid.UUID
	EventID        uuid.UUID
	OrganizationID uuid.UUID
	ActorUserID    *uuid.UUID
	EntityType     string
	EntityID       uuid.UUID
	Action         string
	Changes        json.RawMessage
	CreatedAt      time.Time
}
