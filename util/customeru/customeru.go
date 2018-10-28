package customeru

import (
	"github.com/havr/customers/managers"
	"github.com/havr/customers/models"
	"github.com/icrowley/fake"
	"math/rand"
	"time"
)

// RandomCustomer generates a valid customer with random life-like data
func RandomCustomer() models.Customer {
	customer := models.Customer{
		BirthDate: models.Date(fakeBirthday(managers.MinCustomerAge, managers.MaxCustomerAge).UTC()),
		Email:     fake.EmailAddress(),
		Address:   fake.StreetAddress(),
	}
	isFemale := rand.Int31n(2) == 0
	if isFemale {
		customer.Gender = models.Female
		customer.LastName = fake.FemaleLastName()
		customer.FirstName = fake.FemaleFirstName()
	} else {
		customer.Gender = models.Male
		customer.LastName = fake.MaleLastName()
		customer.FirstName = fake.MaleFirstName()
	}

	return customer
}

func fakeBirthday(minAge, maxAge int) time.Time {
	now := time.Now()
	from := now.AddDate(-maxAge, 0, 0)
	to := now.AddDate(-minAge, 0, 0)
	return time.Unix(randomInt64(from.Unix(), to.Unix()), 0)
}

func randomInt64(min, max int64) int64 {
	if max <= min {
		return min
	}
	return min + rand.Int63n(max-min+1)
}
