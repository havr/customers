package stores

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/pkg/errors"
	"net/url"
	"strings"
)

const protoPrefix = "postgres://"

func ensureProtoPrefix(str string) string {
	if !strings.HasPrefix(str, protoPrefix) {
		str = protoPrefix + str
	}
	return str
}

// DropDB removes the given database from from the given server
func DropDB(ctx context.Context, dbURLStr, dbName string) error {
	db, err := sql.Open("postgres", ensureProtoPrefix(dbURLStr))
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.ExecContext(ctx, "DROP DATABASE "+dbName)
	return err
}

// PrepareDB creates the given database if it doesn't exists and initializes it
// In all cases it returns connection to a database with valid schema
func PrepareDB(ctx context.Context, dbURLStr string) (_ *sql.DB, gerr error) {
	dbURLStr = ensureProtoPrefix(dbURLStr)
	withoutName, dbName := withoutDbName(dbURLStr)
	db, err := sql.Open("postgres", withoutName)
	if err != nil {
		return nil, errors.Wrapf(err, "open postgres server")
	}
	defer db.Close()

	exists, err := databaseExists(ctx, db, dbName)
	if err != nil {
		return nil, err
	}
	if !exists {
		fmt.Printf("Database %v doesn't exist. Creating... \n", dbName)
		if _, crerr := db.ExecContext(ctx, "CREATE DATABASE "+dbName); crerr != nil {
			return nil, crerr
		}
		defer dropOnError(ctx, db, dbName, &gerr)
	}

	return openDB(ctx, dbURLStr, !exists)
}

func withoutDbName(dbURLStr string) (string, string) {
	dbURL, err := url.Parse(dbURLStr)
	if err != nil {
		panic(err)
	}
	dbName := strings.TrimPrefix(dbURL.Path, "/")
	dbURL.Path = ""
	return dbURL.String(), dbName
}

func dropOnError(ctx context.Context, db *sql.DB, dbName string, err *error) {
	r := recover()
	if r == nil && *err == nil {
		return
	}
	if _, err := db.ExecContext(ctx, "DROP DATABASE "+dbName); err != nil {
		fmt.Println("rollback database creation:", err)
	}
	if r != nil {
		*err = fmt.Errorf("%v", err)
	}
}

func openDB(ctx context.Context, url string, init bool) (*sql.DB, error) {
	db, err := sql.Open("postgres", url)
	if err != nil {
		return nil, err
	}

	if init {
		if _, err := db.ExecContext(ctx, string(schema)); err != nil {
			return nil, err
		}
	}

	return db, nil
}

func databaseExists(ctx context.Context, db *sql.DB, dbName string) (bool, error) {
	row := db.QueryRowContext(ctx, `SELECT 1 FROM pg_database WHERE datname=$1`, dbName)
	var n int
	if err := row.Scan(&n); err == sql.ErrNoRows {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}
