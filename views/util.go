package views

import (
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
)

func (v *views) executeTemplate(w http.ResponseWriter, name string, data interface{}) {
	if err := v.template.ExecuteTemplate(w, name, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (v *views) id(r *http.Request) int {
	strID := mux.Vars(r)["id"]
	if strID == "" {
		return -1
	}
	intD, err := strconv.Atoi(strID)
	if err != nil {
		fmt.Println("invalid id", strID, ":", err)
		return -1
	}
	return intD
}

func redirect(w http.ResponseWriter, r *http.Request, where string) {
	if where == "" {
		where = r.Header.Get("Referer")
	}
	http.Redirect(w, r, where, http.StatusFound)
}
