package catalog

import (
	"database/sql"
	"errors"
	"log/slog"
	"strings"
	"uuid"
)

type repository interface {
	Create(uuid.UUID, *uuid.UUID, CreateItemRequest) (Item, error)
	List(uuid.UUID) ([]Item, error)
	FindByID(uuid.UUID, uuid.UUID) (Item, error)
}

type Service struct{ repo repository }

var (
	InvalidKindErr     = errors.New("the item kind is invalid")
	InvalidPriceErr    = errors.New("the item price cannot be negative")
	InvalidCurrencyErr = errors.New("the item currency is invalid")
	ItemConflictErr    = errors.New("an item with this name already exists")
	ItemNotFoundErr    = errors.New("item not found")
	ItemInternalErr    = errors.New("an internal error has occurred")
)

func NewService(repo Repo) *Service { return &Service{repo: repo} }

func (s Service) Create(organizationID uuid.UUID, actorUserID *uuid.UUID, request CreateItemRequest) (Item, error) {
	if request.Kind != ProductKind && request.Kind != ServiceKind {
		return Item{}, InvalidKindErr
	}
	if request.DefaultPriceCents < 0 {
		return Item{}, InvalidPriceErr
	}
	request.Currency = strings.ToUpper(strings.TrimSpace(request.Currency))
	if request.Currency == "" {
		request.Currency = "BRL"
	}
	if len(request.Currency) != 3 {
		return Item{}, InvalidCurrencyErr
	}
	item, err := s.repo.Create(organizationID, actorUserID, request)
	if err != nil {
		if errors.Is(err, ItemConflictErr) {
			return Item{}, err
		}
		slog.Error("catalog.Create", "err", err)
		return Item{}, ItemInternalErr
	}
	return item, nil
}

func (s Service) List(organizationID uuid.UUID) ([]Item, error) {
	items, err := s.repo.List(organizationID)
	if err != nil {
		slog.Error("catalog.List", "err", err)
		return nil, ItemInternalErr
	}
	return items, nil
}

func (s Service) FindByID(organizationID, itemID uuid.UUID) (Item, error) {
	item, err := s.repo.FindByID(organizationID, itemID)
	if errors.Is(err, sql.ErrNoRows) {
		return Item{}, ItemNotFoundErr
	}
	if err != nil {
		slog.Error("catalog.FindByID", "err", err)
		return Item{}, ItemInternalErr
	}
	return item, nil
}
