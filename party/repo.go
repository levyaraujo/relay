package party

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

func (r Repo) Create(organizationID uuid.UUID, actorUserID *uuid.UUID, request CreatePartyRequest) (Party, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return Party{}, err
	}
	defer tx.Rollback()

	var party Party
	party.OrganizationID = organizationID
	party.Name = request.Name
	party.Email = request.Email
	party.Phone = request.Phone
	party.Document = request.Document
	err = tx.QueryRow(`
		INSERT INTO parties (organization_id, name, email, phone, document)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id`,
		organizationID, request.Name, request.Email, request.Phone, request.Document,
	).Scan(&party.ID)
	if err != nil {
		return Party{}, mapRepositoryError(err)
	}

	party.Roles = request.Roles
	for _, role := range request.Roles {
		if _, err := tx.Exec(`INSERT INTO party_roles (party_id, role) VALUES ($1, $2) ON CONFLICT DO NOTHING`, party.ID, role); err != nil {
			return Party{}, err
		}
	}

	payload, err := json.Marshal(party)
	if err != nil {
		return Party{}, err
	}
	if err := r.audit.Enqueue(tx, audit.NewEvent(organizationID, actorUserID, "PARTY_CREATED", "PARTY", party.ID, "CREATE", payload)); err != nil {
		return Party{}, err
	}

	if err := tx.Commit(); err != nil {
		return Party{}, err
	}
	return party, nil
}

func (r Repo) List(organizationID uuid.UUID, role *Role) ([]Party, error) {
	query := `
		SELECT p.id, p.organization_id, p.name, p.email, p.phone, p.document, pr.role
		FROM parties p
		JOIN party_roles pr ON pr.party_id = p.id
		WHERE p.organization_id = $1`
	args := []any{organizationID}
	if role != nil {
		query += ` AND pr.role = $2`
		args = append(args, *role)
	}
	query += ` ORDER BY p.name, p.id, pr.role`

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	parties := make([]Party, 0)
	indexes := make(map[uuid.UUID]int)
	for rows.Next() {
		var party Party
		var role Role
		if err := rows.Scan(&party.ID, &party.OrganizationID, &party.Name, &party.Email, &party.Phone, &party.Document, &role); err != nil {
			return nil, err
		}
		index, ok := indexes[party.ID]
		if !ok {
			party.Roles = []Role{role}
			indexes[party.ID] = len(parties)
			parties = append(parties, party)
			continue
		}
		parties[index].Roles = append(parties[index].Roles, role)
	}
	return parties, rows.Err()
}

func (r Repo) FindByID(organizationID, partyID uuid.UUID) (Party, error) {
	parties, err := r.find(organizationID, &partyID)
	if err != nil {
		return Party{}, err
	}
	if len(parties) == 0 {
		return Party{}, sql.ErrNoRows
	}
	return parties[0], nil
}

func (r Repo) find(organizationID uuid.UUID, partyID *uuid.UUID) ([]Party, error) {
	query := `
		SELECT p.id, p.organization_id, p.name, p.email, p.phone, p.document, pr.role
		FROM parties p
		JOIN party_roles pr ON pr.party_id = p.id
		WHERE p.organization_id = $1 AND p.id = $2
		ORDER BY pr.role`
	rows, err := r.db.Query(query, organizationID, *partyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var parties []Party
	var party Party
	for rows.Next() {
		var role Role
		if err := rows.Scan(&party.ID, &party.OrganizationID, &party.Name, &party.Email, &party.Phone, &party.Document, &role); err != nil {
			return nil, err
		}
		party.Roles = append(party.Roles, role)
	}
	if len(party.Roles) > 0 {
		parties = append(parties, party)
	}
	return parties, rows.Err()
}

func mapRepositoryError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return PartyConflictErr
	}
	return err
}

func CreateRepo(db *sql.DB) Repo {
	return Repo{db: db, audit: audit.CreateRepo(db)}
}
