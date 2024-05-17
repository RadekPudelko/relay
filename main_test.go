package main

import (
	// "bytes"
	// "net/http"
	// "net/http/httptest"
	// "net/url"
	// "encoding/csv"
	// "encoding/json"
	// "fmt"
	// "os"
	// "strconv"

	"database/sql"
	"testing"
	"time"
    "fmt"

	"pcfs/db"
)

func testCreateTask(dbConn *sql.DB, somId string, productId int, cloudFunction string, argument string, desiredReturnCode sql.NullInt32, scheduledTime time.Time) (error) {
    id, err := CreateTask(dbConn, somId, productId, cloudFunction, argument, desiredReturnCode, scheduledTime)
    if err != nil {
		return fmt.Errorf("TestGetTasks: %w", err)
    }

    task, err := db.SelectTask(dbConn, id)
    if err != nil {
		return fmt.Errorf("TestGetTasks: %w", err)
    }

    if task.Som.SomId != somId {
		return fmt.Errorf("TestGetTasks: SelectTask on task %d got somId %s, expected %s", id, task.Som.SomId, somId)
    }
    if task.Som.ProductId != productId {
		return fmt.Errorf("TestGetTasks: SelectTask on task %d got productId %d, expected %d", id, task.Som.ProductId, productId)
    }
    if task.CloudFunction != cloudFunction {
		return fmt.Errorf("TestGetTasks: SelectTask on task %d got cloudFunction %s, expected %s", id, task.CloudFunction, cloudFunction)
    }
    if task.Argument != argument {
		return fmt.Errorf("TestGetTasks: SelectTask on task %d got argument %s, expected %s", id, task.Argument, argument)
    }
    if task.DesiredReturnCode != desiredReturnCode {
		return fmt.Errorf("TestGetTasks: SelectTask on task %d got desiredReturnCode %+v, expected %+v", id, task.DesiredReturnCode, desiredReturnCode)
    }
    if task.ScheduledTime != scheduledTime {
		return fmt.Errorf("TestGetTasks: SelectTask on task %d got scheduledTime %s, expected %s", id, task.ScheduledTime, scheduledTime)
    }
    return nil
}

func TestCreateTasks(t *testing.T) {
    dbConn, err := db.Connect("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("TestGetTasks: %+v", err)
	}
    defer dbConn.Close()

	err = db.CreateTables(dbConn)
	if err != nil {
		t.Fatalf("TestGetTasks: %+v", err)
	}

    somId := "somid1"
    productId := 0
    cloudFunction := "func1"
    argument := ""
    desiredReturnCode := sql.NullInt32{Int32: 0, Valid: false}
    scheduledTime := time.Now().UTC()

    err = testCreateTask(dbConn, somId, productId, cloudFunction, argument, desiredReturnCode, scheduledTime)
	if err != nil {
		t.Fatalf("TestGetTasks: %+v", err)
	}

    somId = "somid1"
    err = testCreateTask(dbConn, somId, productId, cloudFunction, argument, desiredReturnCode, scheduledTime)
	if err != nil {
		t.Fatalf("TestGetTasks: %+v", err)
	}

    desiredReturnCode = sql.NullInt32{Int32: 0, Valid: true}
    err = testCreateTask(dbConn, somId, productId, cloudFunction, argument, desiredReturnCode, scheduledTime)
	if err != nil {
		t.Fatalf("TestGetTasks: %+v", err)
	}
}

func TestGetReadyTasks(t *testing.T) {
    // TODO:
    // Preload database with tasks, have some tasks ready, complete and failed
    // Request tasks with low limit, high limit, cause it to wrap, etc
    // Check tasks to see if you get whats expected
}
