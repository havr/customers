package views

import (
	"net/http"
)

func (v *views) viewCustomerPage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	data := customerData{
		data: data{
			Title: "View a customer",
		},
	}
	var err error
	data.Customer, err = v.customerManager.GetCustomer(ctx, v.id(r))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	v.executeTemplate(w, "view", data)
}

