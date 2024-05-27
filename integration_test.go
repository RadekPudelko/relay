package main

import (
	"database/sql"
	"testing"
	"time"

	"pcfs/db"
)

func TestIntegration(t *testing.T) {
	dbConn, err := db.Connect("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("TestCreateTasks: %+v", err)
	}
	defer dbConn.Close()

	err = db.CreateTables(dbConn)
	if err != nil {
		t.Fatalf("TestCreateTasks: %+v", err)
	}

	somId := "somid0"
	productId := 0
	cloudFunction := "func1"
	argument := ""
	desiredReturnCode := sql.NullInt64{Int64: 0, Valid: false}
	scheduledTime0 := time.Now().UTC()

	tid, err := testCreateTask(dbConn, somId, productId, cloudFunction, argument, desiredReturnCode, scheduledTime0)
	if err != nil {
		t.Fatalf("TestCreateTasks: %+v", err)
	}
	if tid != 1 {
		t.Fatalf("TestCreateTasks: expected to create task id 1, got %d", tid)
	}

	somId = "somid1"
	tid, err = testCreateTask(dbConn, somId, productId, cloudFunction, argument, desiredReturnCode, scheduledTime0)
	if err != nil {
		t.Fatalf("TestCreateTasks: %+v", err)
	}
	if tid != 2 {
		t.Fatalf("TestCreateTasks: expected to create task id 2, got %d", tid)
	}

	desiredReturnCode = sql.NullInt64{Int64: 0, Valid: true}
	tid, err = testCreateTask(dbConn, somId, productId, cloudFunction, argument, desiredReturnCode, scheduledTime0)
	if err != nil {
		t.Fatalf("TestCreateTasks: %+v", err)
	}
	if tid != 3 {
		t.Fatalf("TestCreateTasks: expected to create task id 3, got %d", tid)
	}
}
