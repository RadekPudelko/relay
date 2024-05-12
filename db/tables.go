package db

import (
	"database/sql"
	"fmt"
)

func SetupTables(db *sql.DB) error {
	err := CreateSomsTable(db)
	if err != nil {
		return fmt.Errorf("SetupDatabase: %w", err)
	}
	err = CreateTasksTable(db)
	if err != nil {
		return fmt.Errorf("SetupDatabase: %w", err)
	}
	return err
}

func CreateSomsTable(db *sql.DB) error {
	const create string = `
         CREATE TABLE IF NOT EXISTS soms (
         id INTEGER PRIMARY KEY AUTOINCREMENT,
         som_id TEXT UNIQUE NOT NULL,
         product_id INTEGER NOT NULL,
         last_online DATETIME,
         last_ping DATETIME
         )`
	_, err := db.Exec(create)
	if err != nil {
		return fmt.Errorf("CreateSomsTable: db.Exec: %w", err)
	}
	return err
}

// TODO: Add scheduled time
func CreateTasksTable(db *sql.DB) error {
	const create string = `
        CREATE TABLE IF NOT EXISTS tasks (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        som_key INTEGER NOT NULL,
        cloud_function TEXT NOT NULL,
        argument TEXT,
        desired_return_code INTEGER,
        status INTEGER NOT NULL,
        tries INTEGER NOT NULL,
        response_code INTEGER,
        response_text TEXT,
        FOREIGN KEY(som_key) REFERENCES soms(id)
        )`
	_, err := db.Exec(create)
	if err != nil {
		return fmt.Errorf("CreateTasksTable: db.Exec: %w", err)
	}
	return err
}

