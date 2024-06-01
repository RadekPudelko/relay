package db

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

func CreateTables(db *sql.DB) error {
	err := CreateDevicesTable(db)
	if err != nil {
		return fmt.Errorf("SetupDatabase: %w", err)
	}
	err = CreateRelaysTable(db)
	if err != nil {
		return fmt.Errorf("SetupDatabase: %w", err)
	}
	return nil
}

func CreateDevicesTable(db *sql.DB) error {
	const query string = `
         CREATE TABLE IF NOT EXISTS devices (
         id INTEGER PRIMARY KEY AUTOINCREMENT,
         device_id TEXT UNIQUE NOT NULL,
         last_online DATETIME NULL,
         last_ping DATETIME NULL
         )`
	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("CreateDevicesTable: db.Exec: %w", err)
	}
	return nil
}

// TODO: Add scheduled time
func CreateRelaysTable(db *sql.DB) error {
	const query string = `
        CREATE TABLE IF NOT EXISTS relays (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        device_key INTEGER NOT NULL,
        cloud_function TEXT NOT NULL,
        argument TEXT NOT NULL,
        desired_return_code INTEGER NULL,
        scheduled_time DATETIME NOT NULL,
        status INTEGER NOT NULL,
        tries INTEGER NOT NULL,
        FOREIGN KEY(device_key) REFERENCES devices(id)
        )`
	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("CreateRelaysTable: db.Exec: %w", err)
	}
	return nil
}
