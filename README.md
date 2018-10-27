This is a test assessment for the Wallester.

It uses go1.11 modules. No web dependencies required.

The easiest way to run is to execute
```bash
go run cmd/customers/customers.go
```
 
Possible flags:
```
    --resources <resource dir>  (defaults to "resoucres")
    --host <host to serve on> (defaults to 0.0.0.0:8080)
    --db <postgres connection URL> (defaults to postgres://postgres:mysecretpassword@localhost:5432/testdb?sslmode=disable)
```
