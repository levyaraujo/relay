package vendors

import (
	"errors"
	"time"

	"github.com/levyaraujo/relay/shared"
)

var (
	ErrNameRequired        = errors.New("vendor name is required")
	ErrInvalidPaymentTerms = errors.New("payment terms cannot be negative")
	ErrInvalidCNPJ         = errors.New("invalid CNPJ")
)

// Controller holds business logic for vendor operations.
type Controller struct {
	repo VendorRepository
}

// NewController returns a Controller wired to the given repository.
func NewController(repo VendorRepository) *Controller {
	return &Controller{repo: repo}
}

// Create validates and persists a new vendor.
func (c *Controller) Create(v *Vendor) (*Vendor, error) {
	if !v.ValidateCNPJ() {
		return nil, ErrInvalidCNPJ
	}
	if v.Name == "" {
		return nil, ErrNameRequired
	}
	if v.PaymentTerms < 0 {
		return nil, ErrInvalidPaymentTerms
	}

	v.Model = shared.NewModel()

	return c.repo.Create(v)
}

// Update validates and persists changes to an existing vendor.
func (c *Controller) Update(v *Vendor) (*Vendor, error) {
	if !v.ValidateCNPJ() {
		return nil, ErrInvalidCNPJ
	}
	if v.Name == "" {
		return nil, ErrNameRequired
	}
	if v.PaymentTerms < 0 {
		return nil, ErrInvalidPaymentTerms
	}
	v.UpdatedAt = time.Now()
	return c.repo.Update(v)
}
