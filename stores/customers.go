package stores

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/havr/customers/models"
)

const (
	// CustomerTable is the name for table that contains customers
	CustomerTable = "customers"
)

var allowedFieldsToOrder = []string{"firstname", "lastname", "birthdate", "gender", "email", "address"}
var selectExpr = func() []string {
	return []string{`SELECT id, xmin, lastname, firstname, birthdate, gender, email, address FROM ` + CustomerTable}
}

// ErrChanged occurs when one tries to update an object that has been modified since initial read
var ErrChanged = fmt.Errorf("the object has been changed")

// NewCustomerStore creates new customer store for the given database connection
func NewCustomerStore(db *sql.DB) CustomerStore {
	return &customerStore{
		db: db,
	}
}

// CustomerStore represents SQL persistence layer for customers
type customerStore struct {
	db *sql.DB
}

// CreateCustomer creates the given customer entry and returns the entry with ID and revision set
func (c *customerStore) CreateCustomer(ctx context.Context, customer models.Customer) (models.Customer, error) {
	query := "INSERT INTO " + CustomerTable + `(lastname, firstname, birthdate, gender, email, address) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id, xmin`
	row := c.db.QueryRowContext(ctx, query, customer.LastName, customer.FirstName, time.Time(customer.BirthDate).UTC(), string(customer.Gender), customer.Email, customer.Address)
	result := customer
	if err := row.Scan(&result.ID, &result.Revision); err != nil {
		return models.Customer{}, errors.Wrapf(err, "create customer")
	}
	return result, nil
}

// CountCustomers returns count of customers that satisfy the given filter
func (c *customerStore) CountCustomers(ctx context.Context, filter CustomerListFilter) (int, error) {
	var count int
	where, args := c.filterWhere(filter, nil)
	query := "SELECT COUNT(*) FROM " + CustomerTable + " " + where
	if err := c.db.QueryRowContext(ctx, query, args...).Scan(&count); err != nil {
		return 0, errors.Wrapf(err, "count customers")
	}
	return count, nil
}

// filterWhere formats a WHERE query part that corresponds the given filter and appends values to filter in query args
func (c *customerStore) filterWhere(filter CustomerListFilter, args []interface{}) (string, []interface{}) {
	resultArgs := args
	var whereConditions []string
	if filter.FirstName != "" {
		resultArgs = append(resultArgs, filter.FirstName+"%")
		whereConditions = append(whereConditions, fmt.Sprintf("firstName ILIKE $%d", len(resultArgs)))
	}
	if filter.LastName != "" {
		resultArgs = append(resultArgs, filter.LastName+"%")
		whereConditions = append(whereConditions, fmt.Sprintf("lastName ILIKE $%d", len(resultArgs)))
	}
	if len(whereConditions) == 0 {
		return "", args
	}
	return "WHERE " + strings.Join(whereConditions, " AND "), resultArgs
}

// ListCustomers returns a list of customers that match the given filter and view options
func (c *customerStore) ListCustomers(ctx context.Context, filter CustomerListFilter, options CustomerViewOptions) ([]models.Customer, error) {
	var ok bool
	for _, field := range allowedFieldsToOrder {
		if strings.ToLower(field) == strings.ToLower(options.OrderBy) {
			ok = true
		}
	}
	if !ok && options.OrderBy != "" {
		return nil, fmt.Errorf("unknown field to order by: %q", options.OrderBy)
	}

	queryStr := selectExpr()
	where, args := c.filterWhere(filter, nil)
	queryStr = append(queryStr, where)
	queryStr = append(queryStr, c.viewOptionsQuery(options)...)
	rows, err := c.db.QueryContext(ctx, strings.Join(queryStr, " "), args...)
	if err != nil {
		return nil, errors.Wrapf(err, "query customer list")
	}
	defer rows.Close()

	var customers []models.Customer
	for rows.Next() {
		result, err := c.scanRow(ctx, rows)
		if err != nil {
			return nil, errors.Wrapf(err, "read customer from database")
		}
		customers = append(customers, result)
	}
	return customers, nil
}

func (c *customerStore) orderQuery(orderBy string, orderDesc bool) string {
	var dir string
	if orderDesc {
		dir = "DESC"
	} else {
		dir = "ASC"
	}
	return fmt.Sprintf("ORDER BY %s %s, ID ASC", orderBy, dir)
}

func (c *customerStore) viewOptionsQuery(options CustomerViewOptions) (queryStr []string) {
	if options.OrderBy != "" {
		queryStr = append(queryStr, c.orderQuery(options.OrderBy, options.OrderDesc))
	}
	if options.Offset != 0 {
		queryStr = append(queryStr, "OFFSET "+strconv.Itoa(options.Offset))
	}
	if options.Limit != 0 {
		queryStr = append(queryStr, "LIMIT "+strconv.Itoa(options.Limit))
	}
	return
}

// UpdateCustomer updates replaces a customer model with the given one based on its ID
func (c *customerStore) UpdateCustomer(ctx context.Context, customer models.Customer) (gerr error) {
	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if gerr != nil {
			tx.Rollback()
		} else if err := tx.Commit(); err != nil {
			gerr = err
		}
	}()
	row := tx.QueryRowContext(ctx, "SELECT xmin FROM "+CustomerTable+" WHERE id = $1", customer.ID)
	var revision int
	if serr := row.Scan(&revision); serr != nil {
		return serr
	}
	if revision != customer.Revision {
		return ErrChanged
	}

	query := "UPDATE " + CustomerTable + ` SET lastname = $1, firstname = $2, birthdate = $3, gender = $4, email = $5, address = $6 WHERE id = $7`
	_, err = tx.ExecContext(ctx, query, customer.LastName, customer.FirstName, time.Time(customer.BirthDate), string(customer.Gender), customer.Email, customer.Address, customer.ID)
	if err != nil {
		return errors.Wrapf(err, "update customer %v", customer.ID)
	}
	return nil
}

// DeleteCustomer deletes a customer by its ID
func (c *customerStore) DeleteCustomer(ctx context.Context, id int) error {
	_, err := c.db.ExecContext(ctx, "DELETE FROM "+CustomerTable+" WHERE id = $1", id)
	return err
}

// GetCustomer returns a customer by its ID
func (c *customerStore) GetCustomer(ctx context.Context, id int) (models.Customer, error) {
	query := append(selectExpr(), "WHERE id = $1")
	customer, err := c.scanRow(ctx, c.db.QueryRowContext(ctx, strings.Join(query, " "), id))
	if err == sql.ErrNoRows {
		err = fmt.Errorf("not found")
	}
	if err != nil {
		return models.Customer{}, errors.Wrapf(err, "get customer %v", id)
	}
	return customer, nil
}

// scanRow helps to scan customer row returned by a database into its structure
func (c *customerStore) scanRow(ctx context.Context, scanner rowScanner) (result models.Customer, _ error) {
	if err := scanner.Scan(&result.ID, &result.Revision, &result.LastName, &result.FirstName, &result.BirthDate, &result.Gender, &result.Email, &result.Address); err != nil {
		return models.Customer{}, err
	}
	result.BirthDate = result.BirthDate.UTC()
	return
}

type rowScanner interface {
	Scan(...interface{}) error
}
