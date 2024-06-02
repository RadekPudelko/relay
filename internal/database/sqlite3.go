package database

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

// TODO: Add database operation retry logic
func Setup(path string, walMode bool) (*sql.DB, error) {
	path += "?cache=shared"

	db, err := Connect(path)
	if err != nil {
		return nil, fmt.Errorf("Setup: %w", err)
	}

	if walMode {
		// https://phiresky.github.io/blog/2020/sqlite-performance-tuning/
		_, err = db.Exec("PRAGMA journal_mode=WAL;")
		if err != nil {
			db.Close()
			return nil, fmt.Errorf("Setup: %w", err)
		}
		var mode string
		err = db.QueryRow("PRAGMA journal_mode;").Scan(&mode)
		if err != nil {
			db.Close()
			return nil, fmt.Errorf("Setup: %w", err)
		}
		if mode != "wal" {
			db.Close()
			return nil, fmt.Errorf("Setup: mode want=wal, got=%s", mode)
		}
	}

	_, err = db.Exec("PRAGMA foreign_keys = ON;")
	if err != nil {
		return nil, fmt.Errorf("Setup: failed to enabled foreign key constraints, %w", err)
	}

    _, err = db.Exec("PRAGMA busy_timeout = 5000;")
    if err != nil {
        return nil, fmt.Errorf("Setup: failed to set the busy_timeout, %w", err)
    }

	err = CreateTables(db)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("Setup: %w", err)
	}
	return db, nil
}

func Connect(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, fmt.Errorf("Connect: sql.Open: %w", err)
	}
	err = db.Ping()
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("Connect: db.Ping: %w", err)
	}
	return db, nil
}
