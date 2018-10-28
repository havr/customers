package managers

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/badoux/checkmail"
	"github.com/bearbin/go-age"
	"github.com/havr/customers/models"
	"github.com/havr/customers/stores"
)

const (
	//MinCustomerAge is a minimal allowed customer age
	MinCustomerAge = 18
	//MaxCustomerAge is a maximal allowed customer age
	MaxCustomerAge = 60
)

var (
	// ErrCustomerTooOld occurs when customer age exceeds allowed MaxCustomerAge
	ErrCustomerTooOld = fmt.Errorf("customer is too old")
	// ErrCustomerTooYoung occurs when customer age is lesser then allowed MaxCustomerAge
	ErrCustomerTooYoung = fmt.Errorf("customer is too young")
	// ErrInvalidEmail occurs when email field is invalid
	ErrInvalidEmail = fmt.Errorf("email has invalid format")
)

// CustomerManager represents business logic related to customer management, such as validation
type CustomerManager struct {
	stores.CustomerStore
}

// NewCustomerManager creates a customer manager that uses the given store
func NewCustomerManager(store stores.CustomerStore) *CustomerManager {
	return &CustomerManager{
		CustomerStore: store,
	}
}

// UpdateCustomer updates the given customer model
func (c *CustomerManager) UpdateCustomer(ctx context.Context, customer models.Customer) error {
	if err := c.ValidateCustomer(customer); err != nil {
		return err
	}
	return c.CustomerStore.UpdateCustomer(ctx, customer)
}

// CreateCustomer creates the given customer model
func (c *CustomerManager) CreateCustomer(ctx context.Context, customer models.Customer) (models.Customer, error) {
	if err := c.ValidateCustomer(customer); err != nil {
		return models.Customer{}, err
	}
	return c.CustomerStore.CreateCustomer(ctx, customer)
}

//MultipleErrors aggregates multiple errors into one
type MultipleErrors []error

func (e MultipleErrors) Error() string {
	var strs []string
	for _, err := range e {
		strs = append(strs, err.Error())
	}
	return strings.Join(strs, "\n")
}

// ValidateCustomer validates the given model and returns all errors it encountered, if any
func (c CustomerManager) ValidateCustomer(customer models.Customer) error {
	var errs []error
	errs = c.validateFirstName(errs, customer.FirstName)
	errs = c.validateLastName(errs, customer.LastName)
	errs = c.validateEmail(errs, customer.Email)
	errs = c.validateAddress(errs, customer.Address)
	errs = c.validateBirthDate(errs, customer.BirthDate)
	errs = c.validateGender(errs, customer.Gender)
	if len(errs) == 0 {
		return nil
	}
	return MultipleErrors(errs)
}

func (c CustomerManager) validateFirstName(errs MultipleErrors, value string) MultipleErrors {
	return c.appendValidationError(errs, "first name", value, 100)
}

func (c CustomerManager) validateLastName(errs MultipleErrors, value string) MultipleErrors {
	return c.appendValidationError(errs, "last name", value, 100)
}

func (c CustomerManager) validateAddress(errs MultipleErrors, value string) MultipleErrors {
	return c.appendValidationError(errs, "address", value, 200)
}

func (c CustomerManager) validateEmail(errs MultipleErrors, value string) MultipleErrors {
	if err := c.validateString("email", value, true, 254); err != nil {
		return append(errs, err)
	}
	if err := checkmail.ValidateFormat(value); err != nil {
		return append(errs, ErrInvalidEmail)
	}
	return errs
}

func (c CustomerManager) validateBirthDate(errs MultipleErrors, birthDate models.Date) MultipleErrors {
	if err := c.validateAgeError(birthDate); err != nil {
		return append(errs, err)
	}
	return errs
}

func (c CustomerManager) validateGender(errs MultipleErrors, value models.Gender) MultipleErrors {
	return c.appendValidationError(errs, "gender", string(value), 0)
}

func (c CustomerManager) validateAgeError(date models.Date) error {
	goDate := time.Time(date)
	if goDate.IsZero() {
		return fmt.Errorf("age is undefined")
	}
	customerAge := age.Age(goDate)
	if customerAge < MinCustomerAge {
		return ErrCustomerTooYoung
	} else if customerAge > MaxCustomerAge {
		return ErrCustomerTooOld
	}
	return nil
}

func (c CustomerManager) appendValidationError(errs MultipleErrors, fieldName string, value string, maxLength int) MultipleErrors {
	if err := c.validateString(fieldName, value, true, maxLength); err != nil {
		return append(errs, err)
	}
	return errs
}

func (c CustomerManager) validateString(fieldName, str string, nonEmpty bool, maxLength int) error {
	if nonEmpty && str == "" {
		return fmt.Errorf("%s is empty", fieldName)
	}
	if maxLength > 0 && len(str) > maxLength {
		return fmt.Errorf("%s is too long: maximum allowed length is %d", fieldName, maxLength)
	}
	return nil
}
