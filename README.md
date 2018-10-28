#### What is it
This is a test assessment for the Wallester.
The application allows you to perform basic CRUD on a list of customers.

A quick overview of features:
* List view
    * filter customers by first letters of their first/last names
    * order by all available columns 
    * delete a customer
    * pagination
    
* Create view
    * validate fields
    
* Update view
    * validate fields
    * don't update if the customer has been already updated by someone else

* View view
    * just view a customer data
    
#### Dependencies
The application uses go1.11 modules. No web dependencies are required.
Resources required for the web application to render are located under the `resources` directory.
The application uses postgres as its database. 

#### Running
The easiest way to run the application is to execute
```bash
go run cmd/customers/customers.go --db your-connection-url
```
And head to the `localhost:8080`

If the specified database doesn't exist, the app creates it automatically.
For the sake of simplicity migration-related stuff is omitted.

Possible flags to tweak:
```
    --resources <resource dir>  (defaults to "resoucres")
    --host <host to serve on> (defaults to 0.0.0.0:8080)
    --db <postgres connection URL> (defaults to postgres://postgres:mysecretpassword@localhost:5432/testdb?sslmode=disable)
```

#### Testing
Just do the following command from the root directory:
```
go test ./...
```

If you want to perform integration tests, set the environment variable:
```
export TEST_DB=<postgres connection url>
```
before the `go test`

Please don't include database name in the connection url, as tests create a database per test.
