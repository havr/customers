package managers_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/havr/customers/managers"
	"github.com/havr/customers/models"
	"github.com/havr/customers/stores"
	"github.com/stretchr/testify/require"
)

var (
	validCustomer = models.Customer{
		FirstName: "First Name",
		LastName:  "Last Name",
		BirthDate: models.Date(time.Now().Add(-30 * 365 * 24 * time.Hour)),
		Gender:    models.Male,
		Email:     "fake@email.com",
		Address:   "Address",
	}
)

func TestManagerCreate(t *testing.T) {
	mgr := managers.NewCustomerManager(fakeCustomerStore{})
	ctx := context.Background()
	_, err := mgr.CreateCustomer(ctx, validCustomer)
	require.NoError(t, err)
}

func TestManagerUpdate(t *testing.T) {
	mgr := managers.NewCustomerManager(fakeCustomerStore{})
	ctx := context.Background()
	require.NoError(t, mgr.UpdateCustomer(ctx, validCustomer))
}

func TestManagerCreateEmpty(t *testing.T) {
	mgr := managers.NewCustomerManager(fakeCustomerStore{})
	ctx := context.Background()
	_, err := mgr.CreateCustomer(ctx, models.Customer{})
	errs := err.(managers.MultipleErrors)
	require.Len(t, errs, 6)
}

func TestManagerUpdateEmpty(t *testing.T) {
	mgr := managers.NewCustomerManager(fakeCustomerStore{})
	ctx := context.Background()
	_, err := mgr.CreateCustomer(ctx, models.Customer{})
	errs := err.(managers.MultipleErrors)
	require.Len(t, errs, 6)
}

func TestManagerCreateInvalidEmail(t *testing.T) {
	mgr := managers.NewCustomerManager(fakeCustomerStore{})
	ctx := context.Background()
	withInvalidEmail := validCustomer
	withInvalidEmail.Email = "invalid"
	_, err := mgr.CreateCustomer(ctx, withInvalidEmail)
	requireOneError(t, managers.ErrInvalidEmail, err)
}

func TestManagerUpdateInvalidEmail(t *testing.T) {
	mgr := managers.NewCustomerManager(fakeCustomerStore{})
	ctx := context.Background()
	withInvalidEmail := validCustomer
	withInvalidEmail.Email = "invalid"
	err := mgr.UpdateCustomer(ctx, withInvalidEmail)
	requireOneError(t, managers.ErrInvalidEmail, err)
}

func TestManagerCreateInvalidAge(t *testing.T) {
	mgr := managers.NewCustomerManager(fakeCustomerStore{})
	ctx := context.Background()
	_, err := mgr.CreateCustomer(ctx, tooYoungCustomer())
	requireOneError(t, managers.ErrCustomerTooYoung, err)

	_, err = mgr.CreateCustomer(ctx, tooOldCustomer())
	requireOneError(t, managers.ErrCustomerTooOld, err)
}

func tooYoungCustomer() models.Customer {
	tooYoung := validCustomer
	tooYoung.BirthDate = models.Date(time.Now().AddDate(-17, 11, 30))
	return tooYoung
}

func tooOldCustomer() models.Customer {
	tooYoung := validCustomer
	tooYoung.BirthDate = models.Date(time.Now().AddDate(-61, 0, 0))
	return tooYoung
}

func TestManagerUpdateInvalidAge(t *testing.T) {
	mgr := managers.NewCustomerManager(fakeCustomerStore{})
	ctx := context.Background()
	err := mgr.UpdateCustomer(ctx, tooYoungCustomer())
	requireOneError(t, managers.ErrCustomerTooYoung, err)

	err = mgr.UpdateCustomer(ctx, tooOldCustomer())
	requireOneError(t, managers.ErrCustomerTooOld, err)
}

func TestManagerCreateTooLong(t *testing.T) {
	mgr := managers.NewCustomerManager(fakeCustomerStore{})
	ctx := context.Background()
	_, err := mgr.CreateCustomer(ctx, customerWithTooLongFields())
	require.Error(t, err)
	merr, ok := err.(managers.MultipleErrors)
	require.True(t, ok)
	require.Len(t, merr, 4)
}

func TestManagerUpdateTooLong(t *testing.T) {
	mgr := managers.NewCustomerManager(fakeCustomerStore{})
	ctx := context.Background()
	_, err := mgr.CreateCustomer(ctx, customerWithTooLongFields())
	require.Error(t, err)
	merr, ok := err.(managers.MultipleErrors)
	require.True(t, ok)
	require.Len(t, merr, 4)
}

func customerWithTooLongFields() models.Customer {
	return models.Customer{
		FirstName: fillStriing(101),
		LastName:  fillStriing(101),
		Email:     fillStriing(255),
		Address:   fillStriing(201),
		Gender:    models.Male,
		BirthDate: validCustomer.BirthDate,
	}
}

func fillStriing(length int) (s string) {
	var b strings.Builder
	for i := 0; i < length; i++ {
		b.WriteString(" ")
	}
	return b.String()
}

func requireOneError(t *testing.T, expect error, err error) {
	require.Error(t, err)
	merr, ok := err.(managers.MultipleErrors)
	require.True(t, ok)
	require.Len(t, merr, 1)
	require.Equal(t, expect, merr[0])
}

type fakeCustomerStore struct{}

func (fakeCustomerStore) CreateCustomer(ctx context.Context, customer models.Customer) (models.Customer, error) {
	return customer, nil
}

func (fakeCustomerStore) CountCustomers(ctx context.Context, filter stores.CustomerListFilter) (int, error) {
	return 0, nil
}

func (fakeCustomerStore) ListCustomers(ctx context.Context, filter stores.CustomerListFilter, options stores.CustomerViewOptions) ([]models.Customer, error) {
	return nil, nil
}

func (fakeCustomerStore) UpdateCustomer(ctx context.Context, customer models.Customer) error {
	return nil
}

func (fakeCustomerStore) DeleteCustomer(ctx context.Context, id int) error {
	return nil
}

func (fakeCustomerStore) GetCustomer(ctx context.Context, id int) (models.Customer, error) {
	return models.Customer{}, nil
}
