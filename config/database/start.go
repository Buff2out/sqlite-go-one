package database

import (
	"database/sql"
)

func StartDB(driver string, DSN string) (*sql.DB, error) {
	db, err := sql.Open(driver, DSN)
	if err != nil {
		return nil, err
	}
	return db, nil
}
