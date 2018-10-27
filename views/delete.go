package views

import (
	"fmt"
	"net/http"
)

func (v *views) deleteCustomer(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	err := v.customerManager.DeleteCustomer(ctx, v.id(r))
	if err != nil {
		fmt.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	redirect(w, r, "")
}
