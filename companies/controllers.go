package companies

import (
	"errors"
	"time"

	"github.com/levyaraujo/relay/accounts"
	"github.com/levyaraujo/relay/shared"
)

var (
	ErrNameRequired = errors.New("company name is required")
	ErrInvalidCNPJ  = errors.New("invalid cnpj")
)

// Controller holds business logic for company operations.
type Controller struct {
	repo CompanyRepository
}

// NewController returns a Controller wired to the given repository.
func NewController(repo CompanyRepository) *Controller {
	return &Controller{repo: repo}
}

// Create validates and persists a new company.
func (c *Controller) Create(co *Company, accRepo accounts.AccountRepository) (*Company, error) {
	if co.Name == "" {
		return nil, ErrNameRequired
	}
	if !shared.ValidateCNPJ(co.CNPJ) {
		return nil, ErrInvalidCNPJ
	}

	co.Model = shared.NewModel()
	newCo, err := c.repo.Create(co)

	if err == nil {
		accounts.SeedDefaultCOA(accRepo, newCo.Id)
	}
	return newCo, err
}

// Update validates and persists changes to an existing company.
func (c *Controller) Update(co *Company) (*Company, error) {
	if co.Name == "" {
		return nil, ErrNameRequired
	}
	if !shared.ValidateCNPJ(co.CNPJ) {
		return nil, ErrInvalidCNPJ
	}
	co.UpdatedAt = time.Now()
	return c.repo.Update(co)
}
