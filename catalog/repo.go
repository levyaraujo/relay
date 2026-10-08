package catalog

import (
	"database/sql"
	"encoding/json"
	"errors"
	"relay/audit"
	"uuid"

	"github.com/jackc/pgx/v5/pgconn"
)

type Repo struct {
	db    *sql.DB
	audit audit.Repo
}

func (r Repo) Create(organizationID uuid.UUID, actorUserID *uuid.UUID, request CreateItemRequest) (Item, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return Item{}, err
	}
	defer tx.Rollback()

	var item Item
	item.OrganizationID = organizationID
	item.Name, item.Description, item.Kind = request.Name, request.Description, request.Kind
	item.DefaultPriceCents, item.Currency = request.DefaultPriceCents, request.Currency
	err = tx.QueryRow(`
		INSERT INTO items (organization_id, name, description, kind, default_price_cents, currency)
		VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
		organizationID, request.Name, request.Description, request.Kind, request.DefaultPriceCents, request.Currency).Scan(&item.ID)
	if err != nil {
		return Item{}, mapRepositoryError(err)
	}
	payload, err := json.Marshal(item)
	if err != nil {
		return Item{}, err
	}
	if err := r.audit.Enqueue(tx, audit.NewEvent(organizationID, actorUserID, "ITEM_CREATED", "ITEM", item.ID, "CREATE", payload)); err != nil {
		return Item{}, err
	}
	if err := tx.Commit(); err != nil {
		return Item{}, err
	}
	return item, nil
}

func (r Repo) List(organizationID uuid.UUID) ([]Item, error) {
	rows, err := r.db.Query(`SELECT id, organization_id, name, description, kind, default_price_cents, currency FROM items WHERE organization_id = $1 ORDER BY name, id`, organizationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]Item, 0)
	for rows.Next() {
		var item Item
		if err := rows.Scan(&item.ID, &item.OrganizationID, &item.Name, &item.Description, &item.Kind, &item.DefaultPriceCents, &item.Currency); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r Repo) FindByID(organizationID, itemID uuid.UUID) (Item, error) {
	var item Item
	err := r.db.QueryRow(`SELECT id, organization_id, name, description, kind, default_price_cents, currency FROM items WHERE organization_id = $1 AND id = $2`, organizationID, itemID).Scan(
		&item.ID, &item.OrganizationID, &item.Name, &item.Description, &item.Kind, &item.DefaultPriceCents, &item.Currency)
	return item, err
}

func mapRepositoryError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return ItemConflictErr
	}
	return err
}

func CreateRepo(db *sql.DB) Repo {
	return Repo{db: db, audit: audit.CreateRepo(db)}
}
