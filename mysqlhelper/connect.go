// Package mysqlhelper provides utilities for connecting to a MySQL database.
package mysqlhelper

import (
	"fmt"
	"log"

	"github.com/jmoiron/sqlx"

	_ "github.com/go-sql-driver/mysql" // MySQL driver
)

// Connect establishes a connection to the MySQL database using the provided DSN (Data Source Name).
func Connect(dsn string) (*sqlx.DB, error) {
	db, err := sqlx.Connect("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MySQL: %w", err)
	}

	return db, nil
}

// CheckConnection checks if the MySQL database connection is healthy.
func CheckConnection(dsn string) error {

	db, err := Connect(dsn)
	if err != nil {
		return err
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Println("Error closing database connection:", err)
		}
	}()
	if err := db.Ping(); err != nil {
		return err
	}

	return nil
}
