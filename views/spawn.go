package views

import (
	"github.com/havr/customers/util/customeru"
	"net/http"
)

func (v *views) handleDataGeneration(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	for i := 0; i < 10; i++ {
		customer := customeru.RandomCustomer()
		if _, err := v.customerManager.CreateCustomer(ctx, customer); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	redirect(w, r, "")
}
