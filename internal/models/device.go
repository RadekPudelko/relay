package models

import (
	"database/sql"
	"fmt"
    "time"
)

type Device struct {
	Id         int          `json:"id"`
	DeviceId   string       `json:"device_id"`
	LastOnline *time.Time `json:"last_online"`
}

func SelectDevice(db *sql.DB, key int) (*Device, error) {
	const query string = `SELECT * FROM devices WHERE id = ?`
	stmt, err := db.Prepare(query)
	if err != nil {
		return nil, fmt.Errorf("SelectDevice: db.Prepare: %w", err)
	}
	defer stmt.Close()

	row := stmt.QueryRow(key)
	var device Device
	err = row.Scan(&device.Id, &device.DeviceId, &device.LastOnline)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("SelectDevice: row.Scan: %w", err)
	}
	return &device, nil
}

func SelectDeviceByDeviceId(db *sql.DB, deviceId string) (*Device, error) {
	const query string = `SELECT * FROM devices WHERE device_id = ?`
	stmt, err := db.Prepare(query)
	if err != nil {
		return nil, fmt.Errorf("SelectDeviceByDeviceId: db.Prepare: %w", err)
	}
	defer stmt.Close()

	var device Device
	err = stmt.QueryRow(deviceId).Scan(&device.Id, &device.DeviceId, &device.LastOnline)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		} else {
			return nil, fmt.Errorf("SelectDeviceByDeviceId: stmt.QueryRow: %w", err)
		}
	}
	return &device, nil
}

func UpdateDevice(db *sql.DB, id int, onlineTime *time.Time) error {
	const query string = `
        UPDATE devices
        SET last_online = ?
        WHERE id = ?
    `
	stmt, err := db.Prepare(query)
	if err != nil {
		return fmt.Errorf("UpdateDevice: db.Prepare: %w", err)
	}
	defer stmt.Close()

	result, err := stmt.Exec(onlineTime, id)
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
	const query string = `INSERT INTO devices (device_id, last_online) VALUES (?, ?)`
	stmt, err := db.Prepare(query)
	if err != nil {
		return 0, fmt.Errorf("InsertDevice: db.Prepare: %w", err)
	}
	defer stmt.Close()

	result, err := stmt.Exec(deviceId, nil)
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
// Returns priamary key for the device
// TODO: This can be 1 sql statement
func InsertOrUpdateDevice(db *sql.DB, deviceId string) (int, error) {
	device, err := SelectDeviceByDeviceId(db, deviceId)
	if err != nil {
		return -1, fmt.Errorf("InsertOrUpdateDevice: %w", err)
	}
	if device != nil {
		return device.Id, nil
	}
	return InsertDevice(db, deviceId)
}
