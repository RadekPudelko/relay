package main

import (
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"pcfs/db"
	"pcfs/server"
)

func testCreateTask(dbConn *sql.DB, somId string,  cloudFunction string, argument string, desiredReturnCode sql.NullInt64, scheduledTime0 time.Time) (int, error) {
	id, err := server.CreateTask(dbConn, somId, cloudFunction, argument, desiredReturnCode, scheduledTime0)
	if err != nil {
		return 0, fmt.Errorf("testCreateTask: %w", err)
	}

	task, err := db.SelectTask(dbConn, id)
	if err != nil {
		return 0, fmt.Errorf("testCreateTask: %w", err)
	}

	if task.Som.SomId != somId {
		return 0, fmt.Errorf("testCreateTask: SelectTask on task %d got somId %s, expected %s", id, task.Som.SomId, somId)
	}
	if task.CloudFunction != cloudFunction {
		return 0, fmt.Errorf("testCreateTask: SelectTask on task %d got cloudFunction %s, expected %s", id, task.CloudFunction, cloudFunction)
	}
	if task.Argument != argument {
		return 0, fmt.Errorf("testCreateTask: SelectTask on task %d got argument %s, expected %s", id, task.Argument, argument)
	}
	if task.DesiredReturnCode != desiredReturnCode {
		return 0, fmt.Errorf("testCreateTask: SelectTask on task %d got desiredReturnCode %+v, expected %+v", id, task.DesiredReturnCode, desiredReturnCode)
	}
	if task.ScheduledTime != scheduledTime0 {
		return 0, fmt.Errorf("testCreateTask: SelectTask on task %d got scheduledTime0 %s, expected %s", id, task.ScheduledTime, scheduledTime0)
	}
	return id, nil
}

func TestCreateTasks(t *testing.T) {
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
	cloudFunction := "func1"
	argument := ""
	desiredReturnCode := sql.NullInt64{Int64: 0, Valid: false}
	scheduledTime0 := time.Now().UTC()

	tid, err := testCreateTask(dbConn, somId, cloudFunction, argument, desiredReturnCode, scheduledTime0)
	if err != nil {
		t.Fatalf("TestCreateTasks: %+v", err)
	}
	if tid != 1 {
		t.Fatalf("TestCreateTasks: expected to create task id 1, got %d", tid)
	}

	somId = "somid1"
	tid, err = testCreateTask(dbConn, somId, cloudFunction, argument, desiredReturnCode, scheduledTime0)
	if err != nil {
		t.Fatalf("TestCreateTasks: %+v", err)
	}
	if tid != 2 {
		t.Fatalf("TestCreateTasks: expected to create task id 2, got %d", tid)
	}

	desiredReturnCode = sql.NullInt64{Int64: 0, Valid: true}
	tid, err = testCreateTask(dbConn, somId,  cloudFunction, argument, desiredReturnCode, scheduledTime0)
	if err != nil {
		t.Fatalf("TestCreateTasks: %+v", err)
	}
	if tid != 3 {
		t.Fatalf("TestCreateTasks: expected to create task id 3, got %d", tid)
	}
}

func createCustomTask(dbConn *sql.DB, somId string, cloudFunction, argument string, desiredReturnCode sql.NullInt64, timeStr string, status db.TaskStatus) (int, error) {
	layout := "2006-01-02 15:04:05.999999-07:00"
	scheduledTime, err := time.Parse(layout, timeStr)
	if err != nil {
		return 0, fmt.Errorf("createCustomTask: time.Parse on %s", timeStr)
	}
	scheduledTime = scheduledTime.UTC()

	tid, err := testCreateTask(dbConn, somId, cloudFunction, argument, desiredReturnCode, scheduledTime)
	if err != nil {
		return 0, fmt.Errorf("createCustomTask: %w", err)
	}
	err = db.UpdateTask(dbConn, tid, scheduledTime, status, 1)
	if err != nil {
		return 0, fmt.Errorf("createCustomTask: %w", err)
	}
	return tid, nil
}

func sliceCompare(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func testGetReadyTasks(dbConn *sql.DB, searchTime string, startId, limit int, expectedTasksIds []int) error {
	const layout string = "2006-01-02 15:04:05.999999-07:00"
	// Search all tasks
	testScheduledTime, err := time.Parse(layout, searchTime)
	if err != nil {
		return fmt.Errorf("testGetReadyTasks: time.Parse %w on %s", err, searchTime)
	}
	testScheduledTime = testScheduledTime.UTC()
	tasks, err := server.GetReadyTasks(dbConn, startId, limit, testScheduledTime)
	if err != nil {
		return fmt.Errorf("testGetReadyTasks: %w", err)
	}
	if !sliceCompare(tasks, expectedTasksIds) {
		return fmt.Errorf("TestGetReadyTasks: mismatch tasks, expected %+v, got %+v", expectedTasksIds, tasks)
	}
	return nil
}

func TestGetReadyTasks(t *testing.T) {
	// TODO:
	// Preload database with tasks, have some tasks ready, complete and failed
	// Request tasks with low limit, high limit, cause it to wrap, etc
	// Check tasks to see if you get whats expected
	// dbConn, err := db.Connect("file::memory:?cache=shared")
	testDBPath := "test.db3"
	// Attempt to delete the file
	err := os.Remove(testDBPath)
	if err != nil {
		t.Fatalf("TestGetReadyTasks: %+v", err)
	}
	dbConn, err := db.Connect(testDBPath + "?cache=shared")
	if err != nil {
		t.Fatalf("TestGetReadyTasks: %+v", err)
	}
	defer dbConn.Close()

	err = db.CreateTables(dbConn)
	if err != nil {
		t.Fatalf("TestGetReadyTasks: %+v", err)
	}

	testTimeStr := "2024-05-15 20:17:32.897647+00:00" // 1 day after scheduled tasks

	somId := "som0"
	cloudFunction := "func0"
	argument := ""
	desiredReturnCode := sql.NullInt64{Int64: 0, Valid: false}
	timeStr0 := "2024-05-14 20:17:32.897647+00:00"

	// No tasks
	err = testGetReadyTasks(dbConn, timeStr0, 0, 10, []int{})
	if err != nil {
		t.Fatalf("TestGetReadyTasks: %+v", err)
	}

	// som0
	t0, err := createCustomTask(dbConn, somId, cloudFunction, argument, desiredReturnCode, timeStr0, db.TaskReady)
	if err != nil {
		t.Fatalf("TestGetReadyTasks: %+v", err)
	}
	err = testGetReadyTasks(dbConn, testTimeStr, 1, 1, []int{t0})
	if err != nil {
		t.Fatalf("TestGetReadyTasks: %+v", err)
	}

	// som1
	somId = "som1"
	t1, err := createCustomTask(dbConn, somId, cloudFunction, argument, desiredReturnCode, timeStr0, db.TaskReady)
	if err != nil {
		t.Fatalf("TestGetReadyTasks: %+v", err)
	}
	err = testGetReadyTasks(dbConn, testTimeStr, 1, 1, []int{t0})
	if err != nil {
		t.Fatalf("TestGetReadyTasks: %+v", err)
	}
	err = testGetReadyTasks(dbConn, testTimeStr, 1, 2, []int{t0, t1})
	if err != nil {
		t.Fatalf("TestGetReadyTasks: %+v", err)
	}
	// err = testGetReadyTasks(dbConn, testTimeStr, 2, 3, []int{t1, t0}) // TODO: support wrap
	err = testGetReadyTasks(dbConn, testTimeStr, 2, 3, []int{t1})
	if err != nil {
		t.Fatalf("TestGetReadyTasks: %+v", err)
	}

	// som1, func1
	cloudFunction = "func1"
	t2, err := createCustomTask(dbConn, somId, cloudFunction, argument, desiredReturnCode, timeStr0, db.TaskReady)
	if err != nil {
		t.Fatalf("TestGetReadyTasks: %+v", err)
	}
	err = testGetReadyTasks(dbConn, testTimeStr, 1, 3, []int{t0, t1})
	if err != nil {
		t.Fatalf("TestGetReadyTasks: %+v", err)
	}

	// som1, func1, complete
	t3, err := createCustomTask(dbConn, somId, cloudFunction, argument, desiredReturnCode, timeStr0, db.TaskComplete)
	if err != nil {
		t.Fatalf("TestGetReadyTasks: %+v", err)
	}
	err = testGetReadyTasks(dbConn, testTimeStr, 1, 3, []int{t0, t1})
	if err != nil {
		t.Fatalf("TestGetReadyTasks: %+v", err)
	}

	// som0, func1, failed
	somId = "som0"
	t4, err := createCustomTask(dbConn, somId, cloudFunction, argument, desiredReturnCode, timeStr0, db.TaskFailed)
	if err != nil {
		t.Fatalf("TestGetReadyTasks: %+v", err)
	}
	err = testGetReadyTasks(dbConn, testTimeStr, 1, 3, []int{t0, t1})
	if err != nil {
		t.Fatalf("TestGetReadyTasks: %+v", err)
	}

	// som0, func1, ready
	t5, err := createCustomTask(dbConn, somId, cloudFunction, argument, desiredReturnCode, timeStr0, db.TaskReady)
	if err != nil {
		t.Fatalf("TestGetReadyTasks: %+v", err)
	}
	err = testGetReadyTasks(dbConn, testTimeStr, 1, 3, []int{t0, t1})
	if err != nil {
		t.Fatalf("TestGetReadyTasks: %+v", err)
	}
	err = testGetReadyTasks(dbConn, testTimeStr, t1, 3, []int{t1, t5})
	if err != nil {
		t.Fatalf("TestGetReadyTasks: %+v", err)
	}

	// som2, func1, ready
	somId = "som2"
	t6, err := createCustomTask(dbConn, somId, cloudFunction, argument, desiredReturnCode, timeStr0, db.TaskReady)
	if err != nil {
		t.Fatalf("TestGetReadyTasks: %+v", err)
	}
	err = testGetReadyTasks(dbConn, testTimeStr, t0, 10, []int{t0, t1, t6})
	if err != nil {
		t.Fatalf("TestGetReadyTasks: %+v", err)
	}

	// som2, func2, ready
	cloudFunction = "func2"
	t7, err := createCustomTask(dbConn, somId, cloudFunction, argument, desiredReturnCode, timeStr0, db.TaskReady)
	if err != nil {
		t.Fatalf("TestGetReadyTasks: %+v", err)
	}
	err = testGetReadyTasks(dbConn, testTimeStr, t0, 10, []int{t0, t1, t6})
	if err != nil {
		t.Fatalf("TestGetReadyTasks: %+v", err)
	}

	// som0, func2, ready, in the future
	somId = "som3"
	timeStr1 := "2025-05-14 20:17:32.897647+00:00"
	t8, err := createCustomTask(dbConn, somId, cloudFunction, argument, desiredReturnCode, timeStr1, db.TaskReady)

	if err != nil {
		t.Fatalf("TestGetReadyTasks: %+v", err)
	}
	err = testGetReadyTasks(dbConn, testTimeStr, t0, 10, []int{t0, t1, t6})
	if err != nil {
		t.Fatalf("TestGetReadyTasks: %+v", err)
	}
	t.Log(t0, t1, t2, t3, t4, t5, t6, t7, t8)
}

