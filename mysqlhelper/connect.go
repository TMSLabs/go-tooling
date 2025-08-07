// Package mysqlhelper provides utilities for connecting to a MySQL database.
package mysqlhelper

import (
	"log"
	"os"

	"github.com/jmoiron/sqlx"
)

// CheckConnection checks if the MySQL database connection is healthy.
func CheckConnection() error {

	db, err := sqlx.Connect("mysql", os.Getenv("MYSQL"))
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
