package db

import (
	"database/sql"
	"fmt"
    "time"
)

type Task struct {
	Id                int            `json:"id"`
	Som               *Som           `json:"som"`
	CloudFunction     string         `json:"cloud_function"`
	Argument          string         `json:"argument"`
	DesiredReturnCode sql.NullInt32  `json:"desired_return_code"`
    ScheduledTime time.Time `json:"scheduled_time"`
	Status            TaskStatus     `json:"status"`
	Tries             int            `json:"tries"`
}

func (t Task) String() string {
    return fmt.Sprintf("task id: %d, som: %s, product:%d, function:%s, argument %s", t.Id, t.Som.SomId, t.Som.ProductId, t.CloudFunction, t.Argument)
}

type TaskStatus int
const (
	TaskReady    TaskStatus = 0
	TaskFailed   TaskStatus = 1
	TaskComplete TaskStatus = 2
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
	const query string = `SELECT * FROM tasks WHERE id = ?`
	stmt, err := db.Prepare(query)
	if err != nil {
		return nil, fmt.Errorf("SelectTask: db.Prepare: %w", err)
	}
	defer stmt.Close()
	// TODO: Apply this approach to other single row reads
	row := stmt.QueryRow(id)
	var task Task
	var somKey int
	err = row.Scan(&task.Id, &somKey, &task.CloudFunction, &task.Argument, 
        &task.DesiredReturnCode, &task.ScheduledTime, &task.Status, &task.Tries)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("SelectTask: row.Scan: %w", err)
	}

	task.Som, err = SelectSom(db, somKey)
	if err != nil {
		return nil, fmt.Errorf("SelectTask: %w", err)
	}
	return &task, nil
}

func SelectTaskIds(db *sql.DB, status TaskStatus, id int, scheduledTime time.Time, limit int) ([]int, error) {
	const query string = `
        SELECT MIN(id) AS id
        FROM tasks
        WHERE status = ?
        AND id > ?
        AND scheduled_time < ?
        ORDER BY id ASC
        LIMIT ?
        `
	stmt, err := db.Prepare(query)
	if err != nil {
		return nil, fmt.Errorf("SelectTaskIds: db.Prepare: %w", err)
	}
	defer stmt.Close()
	rows, err := stmt.Query(status, id, scheduledTime, limit)
	// Might have no rows, where does that error pop?
	if err != nil {
		return nil, fmt.Errorf("SelectTaskIds: stmt.Query: %w", err)
	}
	defer rows.Close()

	var taskIds []int
    if !rows.Next() {
        return taskIds, nil
    }

    // SELECT MIN will return a null row if there aren't any tasks instead of 0 rows
    var taskId sql.NullInt32
    if err := rows.Scan(&taskId); err != nil {
        return nil, fmt.Errorf("SelectTaskIds: first row stmt.Query: %w", err)
    }
    // First row is NULL, so there are no tasks
    if !taskId.Valid {
        return taskIds, nil
    }

    // There are tasks
    taskIds = append(taskIds, int(taskId.Int32))
	for rows.Next() {
		var taskId int
		if err := rows.Scan(&taskId); err != nil {
			return nil, fmt.Errorf("SelectTaskIds: rows.Scan: %w", err)
		}
		taskIds = append(taskIds, taskId)
	}
	// Is this necessary?
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("SelectTaskIds: rows.Err: %w", err)
	}
	return taskIds, nil
}

func InsertTask(db *sql.DB, somKey int, cloudFunction string, argument string, desiredReturnCode *int, scheduledTime time.Time) (int, error) {
	const query string = `
        INSERT INTO tasks 
        (som_key, cloud_function, argument, desired_return_code, scheduled_time, status, tries) 
        VALUES (?, ?, ?, ?, ?, ?, ?)
        `
	stmt, err := db.Prepare(query)
	if err != nil {
		return 0, fmt.Errorf("InsertTask: db.Prepare: %w", err)
	}
	result, err := stmt.Exec(somKey, cloudFunction, argument, desiredReturnCode, scheduledTime, TaskReady, 0)
	if err != nil {
		return 0, fmt.Errorf("InsertTask: stmt.Exec: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("InsertTask: result.LastInsertIdId: %w", err)
	}
	return int(id), nil
}

func UpdateTask(db *sql.DB, taskId int, scheduledTime time.Time, status TaskStatus, tries int, ) (error) {
	const query string = `
        UPDATE tasks 
        SET status = ?, tries = ?, scheduled_time = ?
        WHERE id = ?
        `
	stmt, err := db.Prepare(query)
	if err != nil {
		return fmt.Errorf("UpdateTask: db.Prepare: %w", err)
	}
	result, err := stmt.Exec(status, tries, scheduledTime, taskId) 
	if err != nil {
		return fmt.Errorf("UpdateTask: stmt.Exec: %w", err)
	}

    // Is this necessary?
    rows, err := result.RowsAffected()
    if err != nil {
        fmt.Println("Error getting rows affected:", err)
		return fmt.Errorf("UpdateTask: result.RowsAffected: %w", err)
    }
    if rows != 1 {
		return fmt.Errorf("UpdateTask: expected update to affect 1 row, affected %d rows", rows)
    }
	return nil
}

