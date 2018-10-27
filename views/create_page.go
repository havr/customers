package views

import (
	"fmt"
	"github.com/havr/customers/models"
	"github.com/pkg/errors"
	"html/template"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type customerData struct {
	data
	Edit     bool
	Customer models.Customer
}

func (v *views) createCustomerPage(w http.ResponseWriter, r *http.Request) {
	viewData := customerData{
		data: data{
			Title: "Create a Customer",
		},
		Customer: models.Customer{
			BirthDate: models.Date(time.Now().Add(-time.Duration(365*30) * 24 * time.Hour)),
			Gender:    models.Female,
		},
	}
	if r.Method == http.MethodPost {
		customer, err := v.getCustomer(r)
		if err == nil {
			_, err = v.customerManager.CreateCustomer(r.Context(), customer)
		}
		if err == nil {
			redirect(w, r, "/ui/customer/list")
			return
		}
		viewData.Error = v.formatErrorHTML(err)
		viewData.Customer = customer
	}
	v.executeTemplate(w, "create_edit", viewData)
}

func (v *views) formatErrorHTML(err error) template.HTML {
	return template.HTML(strings.Replace(err.Error(), "\n", "<br/>", -1))
}

func (v *views) getCustomer(r *http.Request) (models.Customer, error) {
	var customer models.Customer
	customer.ID = v.id(r)

	strrev := r.FormValue("revision")
	if strrev != "" {
		intrev, err := strconv.Atoi(strrev)
		if err != nil {
			return models.Customer{}, errors.Wrapf(err, "parse revision")
		}
		customer.Revision = intrev
	}
	customer.FirstName = r.FormValue("firstName")
	customer.LastName = r.FormValue("lastName")
	birthDateStr := r.FormValue("birthDate")
	if birthDateStr != "" {
		birthDate, err := time.Parse(jsDateLayout, birthDateStr)
		if err != nil {
			return models.Customer{}, errors.Wrapf(err, "parse birth date")
		}
		customer.BirthDate = models.Date(birthDate)
	}
	customer.Address = r.FormValue("address")
	customer.Email = r.FormValue("email")
	gender := r.FormValue("gender")
	if !models.IsValidGender(gender) {
		return models.Customer{}, fmt.Errorf("unknown gender: %v", gender)
	}
	customer.Gender = models.Gender(gender)
	return customer, nil
}
