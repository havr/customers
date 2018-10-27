package views

import (
	"fmt"
	"github.com/havr/customers/models"
	"github.com/havr/customers/stores"
	"html/template"
	"math"
	"net/http"
	"net/url"
	"strconv"
)

type listData struct {
	data
	stores.CustomerViewOptions
	Filter stores.CustomerListFilter
	Customers []models.Customer
	Pages []page
}

type page struct {
	Title string
	Link string
	Current bool
	Disabled bool
}

const (
	pageSize              = 20
	paginationInnerWindow = 2
)

func (v *views) listCustomersPage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	data := listData{
		data: data{
			Title: "List",
		},
	}
	query := r.URL.Query()
	page, err := v.page(query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	viewOptions := v.viewOptions(query)
	filter := v.getFilter(query)
	total, err := v.customerManager.CountCustomers(ctx, filter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))
	if (totalPages > 0 && page > totalPages) || page < 0 {
		http.Error(w, "invalid page value: " + strconv.Itoa(page), http.StatusInternalServerError)
		return
	}

	viewOptions.Offset = (page - 1) * pageSize
	viewOptions.Limit = pageSize
	if err != nil {
		data.Error = template.HTML(err.Error())
	} else if customers, err := v.customerManager.ListCustomers(ctx, filter, viewOptions); err != nil {
		data.Error = template.HTML(err.Error())
	} else {
		data.Customers = customers
	}
	data.CustomerViewOptions = viewOptions
	data.Filter = filter
	data.Pages = v.makePagination(r.URL.Query() , page, totalPages)
	v.executeTemplate(w, "list", data)
}

func (v *views) pageLink(query url.Values, idx int) string {
	values := make(url.Values)
	for k, v := range query {
		values[k] = v
	}
	values.Set("page", strconv.Itoa(idx))
	return "/ui/customer/list?" + values.Encode()
}

func (v *views) makePagination(query url.Values, current int, totalPages int) []page {
	if totalPages == 0 {
		return nil
	}
	// pages go 1 .. N, rather than 0 .. N - 1
	start := current - paginationInnerWindow
	end := current + paginationInnerWindow
	var pages []page
	pages = append(pages, page{
		Title: "First",
		Link: v.pageLink(query, 1),
		Disabled: current == 1,
	}, page{
		Title: "Previous",
		Link: v.pageLink(query, current - 1),
		Disabled: current == 1,
	})

	if start <= 1 {
		start = 1
	}
	if end >= totalPages {
		end = totalPages
	}
	for i := start; i <= end; i ++ {
		var p page
		p.Title = strconv.Itoa(i)
		if i != current {
			p.Link = v.pageLink(query, i)
		}
		p.Current = i == current
		pages = append(pages, p)
	}
	pages = append(pages, page{
		Title: "Next",
		Link: v.pageLink(query, current + 1),
		Disabled: current >= totalPages,
	}, page{
		Title: "Last",
		Link: v.pageLink(query, totalPages),
		Disabled: current >= totalPages,
	})
	return pages
}

func (v *views) getFilter(query url.Values) (filter stores.CustomerListFilter) {
	filter.FirstName = query.Get("firstName")
	filter.LastName = query.Get("lastName")
	return
}

func (v *views) page(query url.Values) (int, error) {
	pageStr := query.Get("page")
	page := 1
	if pageStr != "" {
		var err error
		if page, err = strconv.Atoi(pageStr); err != nil {
			return 0, fmt.Errorf("invalid page format")
		}
		if page < 0 {
			return 0, fmt.Errorf("invalid page value")
		}
	}
	return page, nil
}

func (v *views) viewOptions(query url.Values) (options stores.CustomerViewOptions) {
	options.OrderBy = query.Get("orderBy")
	if options.OrderBy == "" {
		options.OrderBy = "firstName"
	}
	options.OrderDesc = query.Get("orderDesc") == "true"
	return
}

