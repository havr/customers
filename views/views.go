package views

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/havr/customers/managers"
	"github.com/havr/customers/models"
	"html/template"
	"net/http"
	"path/filepath"
	"time"
)

const jsDateLayout = "2006-01-02"

var funcMap = template.FuncMap{
	"jsDate": func(date models.Date) string {
		return time.Time(date).Format(jsDateLayout)
	},
}

//NewHandler builds a complete http handler for the application
func NewHandler(customerManager *managers.CustomerManager, resourceDir string) http.Handler {
	staticDir := filepath.Join(resourceDir, "static")
	templateDir := filepath.Join(resourceDir, "templates/*.tmpl")
	tmpl, err := template.New("main").Funcs(funcMap).ParseGlob(templateDir)
	if err != nil {
		panic(fmt.Sprintf("process template dir %q: %v", templateDir, err))
	}

	views := &views{
		template:        tmpl,
		customerManager: customerManager,
	}

	router := mux.NewRouter()
	router.Path("/generate").Methods("POST").HandlerFunc(views.handleDataGeneration)

	ui := router.PathPrefix("/ui/customer").Subrouter()
	ui.Path("/list").Methods("GET").HandlerFunc(views.listCustomersPage)
	ui.Path("/create").Methods("GET", "POST").HandlerFunc(views.createCustomerPage)
	ui.Path("/view/{id}").Methods("GET").HandlerFunc(views.viewCustomerPage)
	ui.Path("/edit/{id}").Methods("GET", "POST").HandlerFunc(views.editCustomerPage)
	ui.Path("/delete/{id}").Methods("POST").HandlerFunc(views.deleteCustomer)

	router.Path("/").Methods("GET").Handler(http.RedirectHandler("/ui/customer/list", http.StatusMovedPermanently))
	router.PathPrefix("/static").Handler(http.StripPrefix("/static", http.FileServer(http.Dir(staticDir))))
	return router
}

type views struct {
	template        *template.Template
	customerManager *managers.CustomerManager
}

type data struct {
	Title string
	Error template.HTML
}
