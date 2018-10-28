package views

import (
	"fmt"
	"github.com/havr/customers/stores"
	"net/http"
)

func (v *views) editCustomerPage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	viewData := customerData{
		data: data{
			Title: "Edit a Customer",
		},
		Edit: true,
	}
	if r.Method == http.MethodPost {
		customer, err := v.getCustomer(r)
		if err == nil {
			err = v.customerManager.UpdateCustomer(r.Context(), customer)
			if err == stores.ErrChanged {
				err = fmt.Errorf("somebody has already updated the customer")
				if newcustomer, geterr := v.customerManager.GetCustomer(ctx, customer.ID); geterr != nil {
					err = geterr
				} else {
					customer = newcustomer
				}
			}
		}
		if err == nil {
			redirect(w, r, "/ui/customer/list")
			return
		}

		viewData.Error = v.formatErrorHTML(err)
		viewData.Customer = customer
	} else {
		customer, err := v.customerManager.GetCustomer(ctx, v.id(r))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		viewData.Customer = customer
	}
	v.executeTemplate(w, "create_edit", viewData)
}
