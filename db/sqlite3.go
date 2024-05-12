package db

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

func Connect(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, fmt.Errorf("ConnectToDB: sql.Open: %w", err)
	}
    err = db.Ping()
    if err != nil {
        db.Close()
		return nil, fmt.Errorf("ConnectToDB: db.Ping: %w", err)
    }
	return db, nil
}

