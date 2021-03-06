package stores

import (
	"context"

	"github.com/havr/customers/models"
)

// CustomerViewOptions defines ordering and pagination for list results
type CustomerViewOptions struct {
	OrderBy   string
	OrderDesc bool
	Offset    int
	Limit     int
}

// CustomerListFilter represents filtering options
type CustomerListFilter struct {
	FirstName string
	LastName  string
}

// CustomerStore is a generic interface for customer persistence
type CustomerStore interface {
	CreateCustomer(ctx context.Context, customer models.Customer) (models.Customer, error)
	CountCustomers(ctx context.Context, filter CustomerListFilter) (int, error)
	ListCustomers(ctx context.Context, filter CustomerListFilter, options CustomerViewOptions) ([]models.Customer, error)
	UpdateCustomer(ctx context.Context, customer models.Customer) error
	DeleteCustomer(ctx context.Context, id int) error
	GetCustomer(ctx context.Context, id int) (models.Customer, error)
}
