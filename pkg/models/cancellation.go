package models

import (
	"database/sql"
	"fmt"
)

type Cancellation struct {
	Id      int `json:"id"`
	RelayId int `json:"relay_id"`
}

//TODO: Fix database operations to use Query/Prepare/Exec at appropriate spots
// TODO: add done flag instead of deleting entry?

func SelectCancellations(db *sql.DB, limit int) ([]Cancellation, error) {
	const query string = `SELECT * FROM cancellations LIMIT ?`
	stmt, err := db.Prepare(query)
	if err != nil {
		return nil, fmt.Errorf("SelectCancellation: db.Prepare: %w", err)
	}
	defer stmt.Close()

	rows, err := stmt.Query(limit)
	// Might have no rows, where does that error pop?
	if err != nil {
		return nil, fmt.Errorf("SelectCancellations: stmt.Query: %w", err)
	}
	defer rows.Close()

	var cancellations []Cancellation
	for rows.Next() {
		var cancellation Cancellation
		if err := rows.Scan(&cancellation.Id, &cancellation.RelayId); err != nil {
			return nil, fmt.Errorf("SelectCancellations: rows.Scan: %w", err)
		}
		cancellations = append(cancellations, cancellation)
	}
	// Is this necessary?
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("SelectCancellations: rows.Err: %w", err)
	}
	return cancellations, nil
}

func InsertCancellation(db *sql.DB, relayId int) (int, error) {
	const query string = `INSERT OR IGNORE INTO cancellations (relay_id) VALUES (?)`
	stmt, err := db.Prepare(query)
	if err != nil {
		return 0, fmt.Errorf("InsertCancellation: db.Prepare: %w", err)
	}
	defer stmt.Close()

	result, err := stmt.Exec(relayId)
	if err != nil {
		return 0, fmt.Errorf("InsertCancellation: stmt.Exec: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("InsertCancellation: result.RowsAffected: %w", err)
	}

	if rowsAffected == 0 {
		return 0, nil
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("InsertCancellation: result.LastInsertId: %w", err)
	}

	return int(id), nil
}

func DeleteCancellation(db *sql.DB, id int) error {
	query := `DELETE FROM cancellations WHERE id = ?`
	result, err := db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("DeleteCancellation: db.Exec: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("DeleteCancellation: result.RowsAffected: %w", err)
	}

	if rowsAffected != 1 {
		return fmt.Errorf("DeleteCancellation: rowsAffect want=1, got=%d", rowsAffected)
	}
	return nil
}
