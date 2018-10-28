package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"

	"github.com/havr/customers/managers"
	"github.com/havr/customers/stores"
	"github.com/havr/customers/views"
	_ "github.com/lib/pq"
)

var (
	fResources = flag.String("resources", "resources", "path for application resources")
	fHost      = flag.String("host", "0.0.0.0:8080", "host to serve application")
	fDb        = flag.String("db", "postgres://postgres:mysecretpassword@localhost:5432/testdb?sslmode=disable", "database url")
)

func main() {
	flag.Parse()
	ctx := rootCtx()

	resources := *fResources
	db, err := stores.PrepareDB(ctx, *fDb)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	customerStore := stores.NewCustomerStore(db)
	customerManager := managers.NewCustomerManager(customerStore)

	webLocation := filepath.Join(resources, "web")
	api := views.NewHandler(customerManager, webLocation)
	h := http.Server{
		Addr:    *fHost,
		Handler: api,
	}
	fmt.Println("Serving at", *fHost)
	go func() {
		_ = h.ListenAndServe()
	}()

	<-ctx.Done()
	_ = h.Shutdown(context.Background())
}

func rootCtx() context.Context {
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		done := make(chan os.Signal, 1)
		signal.Notify(done, os.Kill, os.Interrupt)
		<-done
		cancel()
	}()

	return ctx
}
