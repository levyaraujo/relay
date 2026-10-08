package party

import (
	"database/sql"
	"errors"
	"log/slog"
	"relay/shared"
	"uuid"
)

type repository interface {
	Create(uuid.UUID, *uuid.UUID, CreatePartyRequest) (Party, error)
	List(uuid.UUID, *Role) ([]Party, error)
	FindByID(uuid.UUID, uuid.UUID) (Party, error)
}

type Service struct {
	repo repository
}

var (
	InvalidDocumentErr = errors.New("the party document is invalid")
	InvalidRoleErr     = errors.New("the party role is invalid")
	PartyConflictErr   = errors.New("a party with this document already exists")
	PartyNotFoundErr   = errors.New("party not found")
	PartyInternalErr   = errors.New("an internal error has occurred")
)

func (s Service) Create(organizationID uuid.UUID, actorUserID *uuid.UUID, request CreatePartyRequest) (Party, error) {
	request.Document = shared.RemoveSpecialChars(request.Document)
	if !shared.ValidateCNPJ(request.Document) && !shared.ValidateCPF(request.Document) {
		return Party{}, InvalidDocumentErr
	}

	roles, err := normalizeRoles(request.Roles)
	if err != nil {
		return Party{}, err
	}
	request.Roles = roles

	party, err := s.repo.Create(organizationID, actorUserID, request)
	if err != nil {
		if errors.Is(err, PartyConflictErr) {
			return Party{}, err
		}
		slog.Error("party.Create", "err", err)
		return Party{}, PartyInternalErr
	}
	return party, nil
}

func (s Service) List(organizationID uuid.UUID, role *Role) ([]Party, error) {
	parties, err := s.repo.List(organizationID, role)
	if err != nil {
		slog.Error("party.List", "err", err)
		return nil, PartyInternalErr
	}
	return parties, nil
}

func (s Service) FindByID(organizationID, partyID uuid.UUID) (Party, error) {
	party, err := s.repo.FindByID(organizationID, partyID)
	if errors.Is(err, sql.ErrNoRows) {
		return Party{}, PartyNotFoundErr
	}
	if err != nil {
		slog.Error("party.FindByID", "err", err)
		return Party{}, PartyInternalErr
	}
	return party, nil
}

func normalizeRoles(roles []Role) ([]Role, error) {
	if len(roles) == 0 {
		return nil, InvalidRoleErr
	}

	seen := make(map[Role]struct{}, len(roles))
	result := make([]Role, 0, len(roles))
	for _, role := range roles {
		if err := validateRole(role); err != nil {
			return nil, err
		}
		if _, ok := seen[role]; ok {
			continue
		}
		seen[role] = struct{}{}
		result = append(result, role)
	}
	return result, nil
}

func validateRole(role Role) error {
	if role != CustomerRole && role != SupplierRole {
		return InvalidRoleErr
	}
	return nil
}

func NewService(repo Repo) *Service {
	return &Service{repo: repo}
}
