package db

import (
	"database/sql"
	"fmt"
)

type Som struct {
	Id         int          `json:"id"`
	SomId      string       `json:"som_id"`
	ProductId  int          `json:"product_id"`
	LastOnline sql.NullTime `json:"last_online"`
	LastPing   sql.NullTime `json:"last_ping"`
}

func SelectSom(db *sql.DB, key int) (*Som, error) {
	const sel string = `SELECT * FROM soms WHERE id = ?`
	stmt, err := db.Prepare(sel)
	if err != nil {
		return nil, fmt.Errorf("SelectSom: db.Prepare: %w", err)
	}
	defer stmt.Close()

	row, err := stmt.Query(key)
	// Might have no rows, where does that error pop?
	if err != nil {
		return nil, fmt.Errorf("SelectSom: stmt.Query: %w", err)
	}
	defer row.Close()

	var som Som
	err = stmt.QueryRow(key).Scan(&som.Id, &som.SomId, &som.ProductId, &som.LastOnline, &som.LastPing)
	if err != nil {
		return nil, fmt.Errorf("SelectSom: stmt.QueryRow: %w", err)
	}
	return &som, nil
}

func SelectSomBySomId(db *sql.DB, somId string) (*Som, error) {
	const sel string = `SELECT * FROM soms WHERE som_id = ?`
	stmt, err := db.Prepare(sel)
	if err != nil {
		return nil, fmt.Errorf("SelectSomBySomId: db.Prepare: %w", err)
	}
	defer stmt.Close()

	var som Som
	err = stmt.QueryRow(somId).Scan(&som.Id, &som.SomId, &som.ProductId, &som.LastOnline, &som.LastPing)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		} else {
			return nil, fmt.Errorf("SelectSomBySomId: stmt.QueryRow: %w", err)
		}
	}
	return &som, nil
}

func UpdateSom(db *sql.DB, id int, product int, onlineTime, pingTime sql.NullTime) error {
	const update string = `
        UPDATE soms 
        SET product_id = ?, last_online = ?, last_ping = ? 
        WHERE id = ?
    `
	stmt, err := db.Prepare(update)
	if err != nil {
		return fmt.Errorf("UpdateSom: db.Prepare: %w", err)
	}
	defer stmt.Close()

	result, err := stmt.Exec(product, onlineTime, pingTime, id)
	if err != nil {
		return fmt.Errorf("UpdateSom: stmt.Exec: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("UpdateSom: result.rowsAffected: %w", err)
	}
	if rowsAffected != 1 {
		return fmt.Errorf("UpdateSom: expected to affect 1 row, affected %d", rowsAffected)
	}
	return nil
}

func InsertSom(db *sql.DB, somId string, productId int) (int, error) {
	const insert string = `INSERT INTO soms (som_id, product_id, last_online, last_ping) VALUES (?, ?, ?, ?)`
	stmt, err := db.Prepare(insert)
	if err != nil {
		return 0, fmt.Errorf("InsertSom: db.Prepare: %w", err)
	}
	defer stmt.Close()

	result, err := stmt.Exec(somId, productId, nil, nil)
	if err != nil {
		return 0, fmt.Errorf("InsertSom: stmt.Exec: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("InsertSom: result.LastInsertId: %w", err)
	}
	return int(id), nil
}

// Inserts a som into the soms table if it doesn't exist, otherwise updates the productId if
// it doesn't match the one provided.
// Returns priamary key for the som
// TODO: Consider canceling tasks in the event of a productId update
func InsertOrUpdateSom(db *sql.DB, somId string, productId int) (int, error) {
	som, err := SelectSomBySomId(db, somId)
	if err != nil {
		return -1, fmt.Errorf("InsertOrUpdateSom: %w", err)
	}
	if som != nil {
		if som.ProductId == productId {
			return som.Id, nil
		}
		err = UpdateSom(db, som.Id, productId, som.LastOnline, som.LastPing)
		if err != nil {
			return som.Id, fmt.Errorf("InsertOrUpdateSom: %w", err)
		}
	}
	return InsertSom(db, somId, productId)
}
