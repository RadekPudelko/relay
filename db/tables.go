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
	err = CreateTasksTable(db)
	if err != nil {
		return fmt.Errorf("SetupDatabase: %w", err)
	}
	return nil
}

func CreateDevicesTable(db *sql.DB) error {
	const create string = `
         CREATE TABLE IF NOT EXISTS devices (
         id INTEGER PRIMARY KEY AUTOINCREMENT,
         som_id TEXT UNIQUE NOT NULL,
         last_online DATETIME NULL,
         last_ping DATETIME NULL
         )`
	_, err := db.Exec(create)
	if err != nil {
		return fmt.Errorf("CreateSomsTable: db.Exec: %w", err)
	}
	return nil
}

// TODO: Add scheduled time
func CreateTasksTable(db *sql.DB) error {
	const create string = `
        CREATE TABLE IF NOT EXISTS tasks (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        som_key INTEGER NOT NULL,
        cloud_function TEXT NOT NULL,
        argument TEXT NOT NULL,
        desired_return_code INTEGER NULL,
        scheduled_time DATETIME NOT NULL,
        status INTEGER NOT NULL,
        tries INTEGER NOT NULL,
        FOREIGN KEY(som_key) REFERENCES devices(id)
        )`
	_, err := db.Exec(create)
	if err != nil {
		return fmt.Errorf("CreateTasksTable: db.Exec: %w", err)
	}
	return nil
}
