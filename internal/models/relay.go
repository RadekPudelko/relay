package models

import (
	"database/sql"
	"fmt"
	"time"
)

// TODO: decide how to handle nullable vars in marshal/unmarshal
type Relay struct {
	Id                int           `json:"id"`
	Device            *Device       `json:"device"`
	CloudFunction     string        `json:"cloud_function"`
	Argument          string        `json:"argument"`
	DesiredReturnCode sql.NullInt64 `json:"desired_return_code"`
	ScheduledTime     time.Time     `json:"scheduled_time"`
	Status            RelayStatus   `json:"status"`
	Tries             int           `json:"tries"`
}

func (t Relay) String() string {
	return fmt.Sprintf("relay id: %d, device: %s, function:%s, argument %s", t.Id, t.Device.DeviceId, t.CloudFunction, t.Argument)
}

// func NewTask(id int, device *Device, cloudFunction, argument string, desiredReturnCode sql.NullInt64, scheduledTime time.Time, status RelayStatus, tries int) Contact {
//     return Relay {
//         Id: id,
//         Device
//
//
//     }
// 	return Contact{
// 		Id:    id,
// 		Name:  name,
// 		Email: email,
// 	}
// }

type RelayStatus int

const (
	RelayReady    RelayStatus = 0
	RelayFailed   RelayStatus = 1
	RelayComplete RelayStatus = 2
)

// Example of custom field serialization so that instead of reporting sql.NullFields
// as   "response_text": {
//   "String": "",
//   "Valid": false
// }
// they appear as "response_text": null

// type Person struct {
// 	ID           int          `json:"id"`
// 	Name         string       `json:"name"`
// 	Age          int          `json:"age"`
// 	ResponseText NullStringExt `json:"response_text"`
// }
//
// type NullStringExt struct {
// 	sql.NullString
// }
//
// func (n NullStringExt) MarshalJSON() ([]byte, error) {
// 	if !n.Valid {
// 		return []byte("null"), nil
// 	}
// 	return json.Marshal(n.String)
// }

func SelectRelay(models *sql.DB, id int) (*Relay, error) {
	const query string = `SELECT * FROM relays WHERE id = ?`
	stmt, err := models.Prepare(query)
	if err != nil {
		return nil, fmt.Errorf("SelectRelay: models.Prepare: %w", err)
	}
	defer stmt.Close()
	// TODO: Apply this approach to other single row reads
	row := stmt.QueryRow(id)
	var relay Relay
	var deviceKey int
	err = row.Scan(&relay.Id, &deviceKey, &relay.CloudFunction, &relay.Argument,
		&relay.DesiredReturnCode, &relay.ScheduledTime, &relay.Status, &relay.Tries)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("SelectRelay: row.Scan: %w", err)
	}

	relay.Device, err = SelectDevice(models, deviceKey)
	if err != nil {
		return nil, fmt.Errorf("SelectRelay: %w", err)
	}
	return &relay, nil
}

// Select the relays with desired status between with ids betwween start and end (inclusive) occuring after scheduled time.
// Max of 1 taks per device is reutrned (WHERE rn = 1)
func SelectRelayIds(models *sql.DB, status RelayStatus, startId, endId, limit *int, scheduledTime time.Time) ([]int, error) {
	params := []interface{}{status}
	query := `
        SELECT MIN(id)
        FROM relays
        WHERE status = ?
    `
	if startId != nil {
		query += ` AND id >= ?`
		params = append(params, *startId)
	}
	if endId != nil {
		query += ` AND id <= ?`
		params = append(params, *endId)
	}
	query += ` AND scheduled_time <= ?`
	params = append(params, scheduledTime)
	query += ` GROUP BY device_key ORDER by id`
	if limit != nil {
		query += ` LIMIT ?`
		params = append(params, *limit)
	}

	// TODO: figure out how to pretty print these dynamic queries
	// fmt.Println(query)
	// fmt.Println(params)

	stmt, err := models.Prepare(query)
	if err != nil {
		return nil, fmt.Errorf("SelectRelayIds: models.Prepare: %w", err)
	}
	defer stmt.Close()

	rows, err := stmt.Query(params...)
	// Might have no rows, where does that error pop?
	if err != nil {
		return nil, fmt.Errorf("SelectRelayIds: stmt.Query: %w", err)
	}
	defer rows.Close()

	var relayIds []int
	if !rows.Next() {
		return relayIds, nil
	}
	// SELECT MIN will return a null row if there aren't any relays instead of 0 rows
	var relayId sql.NullInt32
	if err := rows.Scan(&relayId); err != nil {
		return nil, fmt.Errorf("SelectRelayIds: first row stmt.Query: %w", err)
	}
	// First row is NULL, so there are no relays
	if !relayId.Valid {
		return relayIds, nil
	}

	// There are relays
	relayIds = append(relayIds, int(relayId.Int32))
	for rows.Next() {
		var relayId int
		if err := rows.Scan(&relayId); err != nil {
			return nil, fmt.Errorf("SelectRelayIds: rows.Scan: %w", err)
		}
		fmt.Println("relay ", relayId)
		relayIds = append(relayIds, relayId)
	}
	// Is this necessary?
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("SelectRelayIds: rows.Err: %w", err)
	}
	return relayIds, nil
}

func InsertRelay(models *sql.DB, deviceKey int, cloudFunction string, argument string, desiredReturnCode sql.NullInt64, scheduledTime time.Time) (int, error) {
	const query string = `
        INSERT INTO relays
        (device_key, cloud_function, argument, desired_return_code, scheduled_time, status, tries)
        VALUES (?, ?, ?, ?, ?, ?, ?)
        `
	stmt, err := models.Prepare(query)
	if err != nil {
		return 0, fmt.Errorf("InsertDevice: models.Prepare: %w", err)
	}
	defer stmt.Close()

	result, err := stmt.Exec(deviceKey, cloudFunction, argument, desiredReturnCode, scheduledTime, RelayReady, 0)
	if err != nil {
		return 0, fmt.Errorf("InsertDevice: stmt.Exec: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("InsertDevice: result.LastInsertIdId: %w", err)
	}
	return int(id), nil
}

func UpdateRelay(models *sql.DB, relayId int, scheduledTime time.Time, status RelayStatus, tries int) error {
	const query string = `
        UPDATE relays
        SET status = ?, tries = ?, scheduled_time = ?
        WHERE id = ?
        `
	stmt, err := models.Prepare(query)
	if err != nil {
		return fmt.Errorf("UpdateRelay: models.Prepare: %w", err)
	}
	defer stmt.Close()

	result, err := stmt.Exec(int(status), tries, scheduledTime, relayId)
	if err != nil {
		return fmt.Errorf("UpdateRelay: stmt.Exec: %w", err)
	}

	// Is this necessary?
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("UpdateRelay: result.RowsAffected: %w", err)
	}
	if rows != 1 {
		return fmt.Errorf("UpdateRelay: expected update to affect 1 row, affected %d rows", rows)
	}
	return nil
}
