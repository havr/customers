package stores_test

import (
	"context"
	"fmt"
	"github.com/havr/customers/models"
	"github.com/havr/customers/stores"
	"github.com/havr/customers/util/customeru"
	"net/url"
	"os"
	"sort"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

func TestCustomers(t *testing.T) {
	dbUrl := os.Getenv("TEST_DB")
	if dbUrl == "" {
		t.Skip("no test database provided")
	}

	tests := map[string]func(t *testing.T, store stores.CustomerStore){
		"list":              tList,
		"listAndCount":      tListAndCount,
		"listAndSort":       tListAndSort,
		"listAndPagination": tListAndPagination,
		"get":               tGet,
		"delete":            tDelete,
		"update":            tUpdate,
		"count":             tCount,
		"filterAndCount":    tFilterAndCount,
	}
	ctx := context.Background()
	for name, test := range tests {
		t.Run(name, func(subt *testing.T) {
			parsed, err := url.Parse(dbUrl)
			if err != nil {
				panic(err)
			}
			dbName := fmt.Sprintf("test%v", time.Now().Nanosecond())
			parsed.Path = dbName
			db, err := stores.PrepareDB(ctx, parsed.String())
			require.NoError(t, err)
			cs := stores.NewCustomerStore(db)
			defer func() {
				_ = db.Close()
				if err := stores.DropDB(ctx, dbUrl, dbName); err != nil {
					fmt.Println("drop db:", err)
				}
			}()

			test(subt, cs)
		})
	}
}

func tUpdate(t *testing.T, store stores.CustomerStore) {
	ctx := context.Background()
	replacements := make(map[models.Customer]bool)
	customers := spawnCustomers(t, ctx, store, 100)
	for customer := range customers {
		replacement := customeru.RandomCustomer()
		replacement.ID = customer.ID
		replacement.Revision = customer.Revision
		replacements[replacement] = true
		require.NoError(t, store.UpdateCustomer(ctx, replacement))

		withFailedRevision := replacement
		withFailedRevision.Revision = customer.Revision - 1
		require.Equal(t, stores.ErrChanged, store.UpdateCustomer(ctx, withFailedRevision))
	}
	for customer := range customers {
		changed, err := store.GetCustomer(ctx, customer.ID)
		require.NoError(t, err)
		changed.Revision = customer.Revision // as its revision has changed
		require.True(t, replacements[changed])
		delete(replacements, changed)
	}
}

func tDelete(t *testing.T, store stores.CustomerStore) {
	ctx := context.Background()
	customers := spawnCustomers(t, ctx, store, 100)
	for customer := range customers {
		require.NoError(t, store.DeleteCustomer(ctx, customer.ID))
		_, err := store.GetCustomer(ctx, customer.ID)
		require.Error(t, err)
	}
	list, err := store.ListCustomers(ctx, stores.CustomerListFilter{}, stores.CustomerViewOptions{})
	require.NoError(t, err)
	require.Len(t, list, 0)
}

func tGet(t *testing.T, store stores.CustomerStore) {
	ctx := context.Background()
	customers := spawnCustomers(t, ctx, store, 100)
	for customer := range customers {
		stored, err := store.GetCustomer(ctx, customer.ID)
		require.NoError(t, err)
		require.Equal(t, customer, stored)
	}
}

func spawnCustomers(t *testing.T, ctx context.Context, store stores.CustomerStore, n int) map[models.Customer]bool {
	customers := make(map[models.Customer]bool)
	for i := 0; i < n; i++ {
		customer, err := store.CreateCustomer(ctx, customeru.RandomCustomer())
		require.NoError(t, err)
		customers[customer] = true
	}
	return customers
}

func tCount(t *testing.T, store stores.CustomerStore) {
	ctx := context.Background()
	customers := spawnCustomers(t, ctx, store, 100)
	count, err := store.CountCustomers(ctx, stores.CustomerListFilter{})
	require.NoError(t, err)
	require.Equal(t, len(customers), count)
}

type filtered struct {
	firstName map[string][]models.Customer
	lastName  map[string][]models.Customer
	comboName map[[2]string][]models.Customer
}

func ditributeFiltered(customers map[models.Customer]bool) (filtered filtered) {
	filtered.firstName = make(map[string][]models.Customer)
	filtered.lastName = make(map[string][]models.Customer)
	filtered.comboName = make(map[[2]string][]models.Customer)
	for customer := range customers {
		fn := string(customer.FirstName[0])
		ln := string(customer.LastName[0])
		filtered.firstName[fn] = append(filtered.firstName[fn], customer)
		filtered.lastName[ln] = append(filtered.lastName[ln], customer)
		comboKey := [2]string{fn, ln}
		filtered.comboName[comboKey] = append(filtered.comboName[comboKey], customer)
	}
	return
}

func tFilterAndCount(t *testing.T, store stores.CustomerStore) {
	ctx := context.Background()
	customers := spawnCustomers(t, ctx, store, 100)
	filtered := ditributeFiltered(customers)
	for fn, expect := range filtered.firstName {
		count, err := store.CountCustomers(ctx, stores.CustomerListFilter{FirstName: fn})
		require.NoError(t, err)
		require.Equal(t, len(expect), count)
	}
	for ln, expect := range filtered.lastName {
		count, err := store.CountCustomers(ctx, stores.CustomerListFilter{LastName: ln})
		require.NoError(t, err)
		require.Equal(t, len(expect), count)
	}
	for combo, expect := range filtered.comboName {
		fn, ln := combo[0], combo[1]
		count, err := store.CountCustomers(ctx, stores.CustomerListFilter{LastName: ln, FirstName: fn})
		require.NoError(t, err)
		require.Equal(t, len(expect), count)
	}
}

func tList(t *testing.T, store stores.CustomerStore) {
	ctx := context.Background()
	customers := spawnCustomers(t, ctx, store, 100)
	entries, err := store.ListCustomers(ctx, stores.CustomerListFilter{}, stores.CustomerViewOptions{})
	require.NoError(t, err)
	for _, entry := range entries {
		require.True(t, customers[entry])
		delete(customers, entry)
	}
	require.Equal(t, 0, len(customers))
}

func sortByFirstName(customers []models.Customer, desc bool) {
	sort.SliceStable(customers, func(i, j int) bool {
		var less bool
		if desc {
			less = customers[i].FirstName > customers[j].FirstName
		} else {
			less = customers[i].FirstName < customers[j].FirstName
		}
		if less {
			return less
		}
		if customers[i].FirstName == customers[j].FirstName {
			return customers[i].ID < customers[j].ID
		}
		return false
	})
}

func sortByDate(customers []models.Customer, desc bool) {
	sort.SliceStable(customers, func(i, j int) bool {
		var less bool
		if desc {
			less = time.Time(customers[i].BirthDate).After(time.Time(customers[j].BirthDate))
		} else {
			less = time.Time(customers[i].BirthDate).Before(time.Time(customers[j].BirthDate))
		}
		if less {
			return less
		}
		if customers[i].FirstName == customers[j].FirstName {
			return customers[i].ID < customers[j].ID
		}
		return false
	})
}

func tListAndCount(t *testing.T, store stores.CustomerStore) {
	ctx := context.Background()
	customers := spawnCustomers(t, ctx, store, 100)
	filtered := ditributeFiltered(customers)
	for fn, list := range filtered.firstName {
		result, err := store.ListCustomers(ctx, stores.CustomerListFilter{FirstName: fn}, stores.CustomerViewOptions{})
		require.NoError(t, err)
		sortByFirstName(result, false)
		sortByFirstName(list, false)
		require.Equal(t, len(list), len(result))
	}
	for ln, list := range filtered.lastName {
		result, err := store.ListCustomers(ctx, stores.CustomerListFilter{LastName: ln}, stores.CustomerViewOptions{})
		require.NoError(t, err)
		sortByFirstName(result, false)
		sortByFirstName(list, false)
		require.Equal(t, len(list), len(result))
	}
	for combo, list := range filtered.comboName {
		fn, ln := combo[0], combo[1]
		result, err := store.ListCustomers(ctx, stores.CustomerListFilter{LastName: ln, FirstName: fn}, stores.CustomerViewOptions{})
		require.NoError(t, err)
		sortByFirstName(result, false)
		sortByFirstName(list, false)
		require.Equal(t, len(list), len(result))
	}
}

func checkSorted(t *testing.T, store stores.CustomerStore, orderField string, orderDesc bool, expect []models.Customer) {
	ctx := context.Background()
	customers, err := store.ListCustomers(ctx, stores.CustomerListFilter{}, stores.CustomerViewOptions{
		OrderBy:   orderField,
		OrderDesc: orderDesc,
	})
	require.NoError(t, err)
	require.Equal(t, expect, customers)
}

func tListAndSort(t *testing.T, store stores.CustomerStore) {
	ctx := context.Background()
	customerMap := spawnCustomers(t, ctx, store, 50)
	var customers []models.Customer
	for customer := range customerMap {
		customers = append(customers, customer)
	}
	sortByFirstName(customers, false)
	checkSorted(t, store, "firstName", false, customers)
	sortByFirstName(customers, true)
	checkSorted(t, store, "firstName", true, customers)

	sortByDate(customers, false)
	checkSorted(t, store, "birthDate", false, customers)
	sortByDate(customers, true)
	checkSorted(t, store, "birthDate", true, customers)
}

func tListAndPagination(t *testing.T, store stores.CustomerStore) {
	ctx := context.Background()
	customerMap := spawnCustomers(t, ctx, store, 50)
	var customers []models.Customer
	for customer := range customerMap {
		customers = append(customers, customer)
	}
	testOffset := len(customers) / 2
	testLimit := len(customers) / 4
	sortByFirstName(customers, false)
	result, err := store.ListCustomers(ctx, stores.CustomerListFilter{}, stores.CustomerViewOptions{
		OrderBy:   "firstName",
		OrderDesc: false,
		Offset:    testOffset,
		Limit:     testLimit,
	})
	require.NoError(t, err)
	require.Equal(t, customers[testOffset:testOffset+testLimit], result)
}
