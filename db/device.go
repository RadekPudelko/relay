package db

import (
	"database/sql"
	"fmt"
)

type Device struct {
	Id         int          `json:"id"`
	DeviceId   string       `json:"device_id"`
	LastOnline sql.NullTime `json:"last_online"`
	LastPing   sql.NullTime `json:"last_ping"`
}

func SelectDevice(db *sql.DB, key int) (*Device, error) {
	const query string = `SELECT * FROM devices WHERE id = ?`
	stmt, err := db.Prepare(query)
	if err != nil {
		return nil, fmt.Errorf("SelectDevice: db.Prepare: %w", err)
	}
	defer stmt.Close()

	row, err := stmt.Query(key)
	// Might have no rows, where does that error pop?
	if err != nil {
		return nil, fmt.Errorf("SelectDevice: stmt.Query: %w", err)
	}
	defer row.Close()

	var dev Device
	err = stmt.QueryRow(key).Scan(&dev.Id, &dev.DeviceId, &dev.LastOnline, &dev.LastPing)
	if err != nil {
		return nil, fmt.Errorf("SelectDevice: stmt.QueryRow: %w", err)
	}
	return &dev, nil
}

func SelectDeviceByDeviceId(db *sql.DB, deviceId string) (*Device, error) {
	const query string = `SELECT * FROM devices WHERE som_id = ?`
	stmt, err := db.Prepare(query)
	if err != nil {
		return nil, fmt.Errorf("SelectDeviceByDeviceId: db.Prepare: %w", err)
	}
	defer stmt.Close()

	var dev Device
	err = stmt.QueryRow(deviceId).Scan(&dev.Id, &dev.DeviceId, &dev.LastOnline, &dev.LastPing)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		} else {
			return nil, fmt.Errorf("SelectDeviceByDeviceId: stmt.QueryRow: %w", err)
		}
	}
	return &dev, nil
}

func UpdateDevice(db *sql.DB, id int, onlineTime, pingTime sql.NullTime) error {
	const query string = `
        UPDATE devices
        SET last_online = ?, last_ping = ?
        WHERE id = ?
    `
	stmt, err := db.Prepare(query)
	if err != nil {
		return fmt.Errorf("UpdateDevice: db.Prepare: %w", err)
	}
	defer stmt.Close()

	result, err := stmt.Exec(onlineTime, pingTime, id)
	if err != nil {
		return fmt.Errorf("UpdateDevice: stmt.Exec: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("UpdateDevice: result.rowsAffected: %w", err)
	}
	if rowsAffected != 1 {
		return fmt.Errorf("UpdateDevice: expected to affect 1 row, affected %d", rowsAffected)
	}
	return nil
}

func InsertDevice(db *sql.DB, deviceId string) (int, error) {
	const query string = `INSERT INTO devices (som_id, last_online, last_ping) VALUES (?, ?, ?)`
	stmt, err := db.Prepare(query)
	if err != nil {
		return 0, fmt.Errorf("InsertDevice: db.Prepare: %w", err)
	}
	defer stmt.Close()

	result, err := stmt.Exec(deviceId, nil, nil)
	if err != nil {
		return 0, fmt.Errorf("InsertDevice: stmt.Exec: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("InsertDevice: result.LastInsertId: %w", err)
	}
	return int(id), nil
}

// Inserts a device into the devices table if it doesn't exist
// Returns priamary key for the dev
// TODO: This can be 1 sql statement
func InsertOrUpdateDevice(db *sql.DB, deviceId string) (int, error) {
	dev, err := SelectDeviceByDeviceId(db, deviceId)
	if err != nil {
		return -1, fmt.Errorf("InsertOrUpdateDevice: %w", err)
	}
	if dev != nil {
		return dev.Id, nil
	}
	return InsertDevice(db, deviceId)
}
