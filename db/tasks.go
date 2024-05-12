package db

import (
	"database/sql"
	"fmt"
)

type Task struct {
	Id                int            `json:"id"`
	Som               *Som           `json:"som"`
	CloudFunction     string         `json:"cloud_function"`
	Argument          string         `json:"argument"`
	DesiredReturnCode sql.NullInt32  `json:"desired_return_code"`
	Status            TaskStatus     `json:"status"`
	Tries             int            `json:"tries"`
	ResponseCode      sql.NullInt16  `json:"resposne_code"`
	ResponseText      sql.NullString `json:"response_text"`
}

func (t Task) String() string {
    return fmt.Sprintf("task id: %d, som: %s, product:%d, function:%s, argument %s", t.Id, t.Som.SomId, t.Som.ProductId, t.CloudFunction, t.Argument)
}

type TaskStatus int

const (
	Ready    TaskStatus = 0
	Failed   TaskStatus = 1
	Complete TaskStatus = 2
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

func SelectTask(db *sql.DB, id int) (*Task, error) {
	const sel string = `SELECT * FROM tasks WHERE id = ?`
	stmt, err := db.Prepare(sel)
	if err != nil {
		return nil, fmt.Errorf("SelTask: db.Prepare: %w", err)
	}
	defer stmt.Close()
	// TODO: Apply this approach to other single row reads
	row := stmt.QueryRow(id)
	var task Task
	var somKey int
	err = row.Scan(&task.Id, &somKey, &task.CloudFunction,
		&task.Argument, &task.DesiredReturnCode, &task.Status, &task.Tries,
		&task.ResponseCode, &task.ResponseText)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("SelTask: row.Scan: %w", err)
	}

	task.Som, err = SelectSomByKey(db, somKey)
	if err != nil {
		return nil, fmt.Errorf("SelTask: %w", err)
	}
	return &task, nil
}

func SelectTaskIds(db *sql.DB, status, id, limit int) ([]int, error) {
	const sel string = `
        SELECT MIN(id) AS id 
        FROM tasks
        WHERE status = ?
        AND id >= ?
        ORDER BY id ASC
        LIMIT ?
        `
	stmt, err := db.Prepare(sel)
	if err != nil {
		return nil, fmt.Errorf("TaskIds: db.Prepare: %w", err)
	}
	defer stmt.Close()
	rows, err := stmt.Query(status, id, limit)
	// Might have no rows, where does that error pop?
	if err != nil {
		return nil, fmt.Errorf("TaskIds: stmt.Query: %w", err)
	}
	defer rows.Close()

	var taskIds []int
	for rows.Next() {
		var taskId int
		if err := rows.Scan(&taskId); err != nil {
			return nil, fmt.Errorf("TaskIds: rows.Scan: %w", err)
		}
		taskIds = append(taskIds, taskId)
	}
	// Is this necessary?
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("TaskIds: rows.Err: %w", err)
	}
	return taskIds, nil
}

func InsertTask(db *sql.DB, somKey int, cloudFunction string, argument *string, desiredReturnCode *int) (int, error) {
	const insert string = `
        INSERT INTO tasks 
        (som_key, cloud_function, argument, desired_return_code, status, tries) 
        VALUES (?, ?, ?, ?, ?, ?)
        `
	stmt, err := db.Prepare(insert)
	if err != nil {
		return 0, fmt.Errorf("InsTask: db.Prepare: %w", err)
	}
	result, err := stmt.Exec(somKey, cloudFunction, argument, desiredReturnCode, Ready, 0)
	if err != nil {
		return 0, fmt.Errorf("InsTask: stmt.Exec: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("InsTask: result.LastInsertIdId: %w", err)
	}
	return int(id), nil
}

