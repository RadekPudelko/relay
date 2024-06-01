package test

import (
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"relay/internal/database"
	"relay/internal/models"
	"relay/internal/server"
)

func testCreateRelay(dbConn *sql.DB, relayId string, cloudFunction string, argument string, desiredReturnCode sql.NullInt64, scheduledTime0 time.Time) (int, error) {
	id, err := server.CreateRelay(dbConn, relayId, cloudFunction, argument, desiredReturnCode, scheduledTime0)
	if err != nil {
		return 0, fmt.Errorf("testCreateRelay: %w", err)
	}

	relay, err := models.SelectRelay(dbConn, id)
	if err != nil {
		return 0, fmt.Errorf("testCreateRelay: %w", err)
	}

	if relay.Device.DeviceId != relayId {
		return 0, fmt.Errorf("testCreateRelay: SelectRelay on relay %d got relayId %s, expected %s", id, relay.Device.DeviceId, relayId)
	}
	if relay.CloudFunction != cloudFunction {
		return 0, fmt.Errorf("testCreateRelay: SelectRelay on relay %d got cloudFunction %s, expected %s", id, relay.CloudFunction, cloudFunction)
	}
	if relay.Argument != argument {
		return 0, fmt.Errorf("testCreateRelay: SelectRelay on relay %d got argument %s, expected %s", id, relay.Argument, argument)
	}
	if relay.DesiredReturnCode != desiredReturnCode {
		return 0, fmt.Errorf("testCreateRelay: SelectRelay on relay %d got desiredReturnCode %+v, expected %+v", id, relay.DesiredReturnCode, desiredReturnCode)
	}
	if relay.ScheduledTime != scheduledTime0 {
		return 0, fmt.Errorf("testCreateRelay: SelectRelay on relay %d got scheduledTime0 %s, expected %s", id, relay.ScheduledTime, scheduledTime0)
	}
	return id, nil
}

func TestCreateRelays(t *testing.T) {
	dbConn, err := database.Connect("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("TestCreateRelays: %+v", err)
	}
	defer dbConn.Close()

	err = database.CreateTables(dbConn)
	if err != nil {
		t.Fatalf("TestCreateRelays: %+v", err)
	}

	relayId := "devid0"
	cloudFunction := "func1"
	argument := ""
	desiredReturnCode := sql.NullInt64{Int64: 0, Valid: false}
	scheduledTime0 := time.Now().UTC()

	tid, err := testCreateRelay(dbConn, relayId, cloudFunction, argument, desiredReturnCode, scheduledTime0)
	if err != nil {
		t.Fatalf("TestCreateRelays: %+v", err)
	}
	if tid != 1 {
		t.Fatalf("TestCreateRelays: expected to create relay id 1, got %d", tid)
	}

	relayId = "devid1"
	tid, err = testCreateRelay(dbConn, relayId, cloudFunction, argument, desiredReturnCode, scheduledTime0)
	if err != nil {
		t.Fatalf("TestCreateRelays: %+v", err)
	}
	if tid != 2 {
		t.Fatalf("TestCreateRelays: expected to create relay id 2, got %d", tid)
	}

	desiredReturnCode = sql.NullInt64{Int64: 0, Valid: true}
	tid, err = testCreateRelay(dbConn, relayId, cloudFunction, argument, desiredReturnCode, scheduledTime0)
	if err != nil {
		t.Fatalf("TestCreateRelays: %+v", err)
	}
	if tid != 3 {
		t.Fatalf("TestCreateRelays: expected to create relay id 3, got %d", tid)
	}
}

func createCustomRelay(dbConn *sql.DB, relayId string, cloudFunction, argument string, desiredReturnCode sql.NullInt64, timeStr string, status models.RelayStatus) (int, error) {
	layout := "2006-01-02 15:04:05.999999-07:00"
	scheduledTime, err := time.Parse(layout, timeStr)
	if err != nil {
		return 0, fmt.Errorf("createCustomRelay: time.Parse on %s", timeStr)
	}
	scheduledTime = scheduledTime.UTC()

	tid, err := testCreateRelay(dbConn, relayId, cloudFunction, argument, desiredReturnCode, scheduledTime)
	if err != nil {
		return 0, fmt.Errorf("createCustomRelay: %w", err)
	}
	err = models.UpdateRelay(dbConn, tid, scheduledTime, status, 1)
	if err != nil {
		return 0, fmt.Errorf("createCustomRelay: %w", err)
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

func testGetReadyRelay(dbConn *sql.DB, searchTime string, startId, limit int, expectedRelayIds []int) error {
	const layout string = "2006-01-02 15:04:05.999999-07:00"
	// Search all relays
	testScheduledTime, err := time.Parse(layout, searchTime)
	if err != nil {
		return fmt.Errorf("testGetReadyRelay: time.Parse %w on %s", err, searchTime)
	}
	testScheduledTime = testScheduledTime.UTC()
	relays, err := server.GetReadyRelays(dbConn, startId, limit, testScheduledTime)
	if err != nil {
		return fmt.Errorf("testGetReadyRelay: %w", err)
	}
	if !sliceCompare(relays, expectedRelayIds) {
		return fmt.Errorf("TestGetReadyRelays: mismatch relays, expected %+v, got %+v", expectedRelayIds, relays)
	}
	return nil
}

func TestGetReadyRelays(t *testing.T) {
	// TODO:
	// Preload database with relays, have some relays ready, complete and failed
	// Request relays with low limit, high limit, cause it to wrap, etc
	// Check relays to see if you get whats expected
	// dbConn, err := db.Connect("file::memory:?cache=shared")
	testDBPath := "test.db3"
	// Attempt to delete the file
	err := os.Remove(testDBPath)
	if err != nil {
		t.Fatalf("TestGetReadyRelays: %+v", err)
	}
	dbConn, err := database.Connect(testDBPath + "?cache=shared")
	if err != nil {
		t.Fatalf("TestGetReadyRelays: %+v", err)
	}
	defer dbConn.Close()

	err = database.CreateTables(dbConn)
	if err != nil {
		t.Fatalf("TestGetReadyRelays: %+v", err)
	}

	testTimeStr := "2024-05-15 20:17:32.897647+00:00" // 1 day after scheduled relays

	relayId := "dev0"
	cloudFunction := "func0"
	argument := ""
	desiredReturnCode := sql.NullInt64{Int64: 0, Valid: false}
	timeStr0 := "2024-05-14 20:17:32.897647+00:00"

	// No relays
	err = testGetReadyRelay(dbConn, timeStr0, 0, 10, []int{})
	if err != nil {
		t.Fatalf("TestGetReadyRelays: %+v", err)
	}

	// dev0
	t0, err := createCustomRelay(dbConn, relayId, cloudFunction, argument, desiredReturnCode, timeStr0, models.RelayReady)
	if err != nil {
		t.Fatalf("TestGetReadyRelays: %+v", err)
	}
	err = testGetReadyRelay(dbConn, testTimeStr, 1, 1, []int{t0})
	if err != nil {
		t.Fatalf("TestGetReadyRelays: %+v", err)
	}

	// dev1
	relayId = "dev1"
	t1, err := createCustomRelay(dbConn, relayId, cloudFunction, argument, desiredReturnCode, timeStr0, models.RelayReady)
	if err != nil {
		t.Fatalf("TestGetReadyRelays: %+v", err)
	}
	err = testGetReadyRelay(dbConn, testTimeStr, 1, 1, []int{t0})
	if err != nil {
		t.Fatalf("TestGetReadyRelays: %+v", err)
	}
	err = testGetReadyRelay(dbConn, testTimeStr, 1, 2, []int{t0, t1})
	if err != nil {
		t.Fatalf("TestGetReadyRelays: %+v", err)
	}
	// err = testGetReadyRelay(dbConn, testTimeStr, 2, 3, []int{t1, t0}) // TODO: support wrap
	err = testGetReadyRelay(dbConn, testTimeStr, 2, 3, []int{t1})
	if err != nil {
		t.Fatalf("TestGetReadyRelays: %+v", err)
	}

	// dev1, func1
	cloudFunction = "func1"
	t2, err := createCustomRelay(dbConn, relayId, cloudFunction, argument, desiredReturnCode, timeStr0, models.RelayReady)
	if err != nil {
		t.Fatalf("TestGetReadyRelays: %+v", err)
	}
	err = testGetReadyRelay(dbConn, testTimeStr, 1, 3, []int{t0, t1})
	if err != nil {
		t.Fatalf("TestGetReadyRelays: %+v", err)
	}

	// dev1, func1, complete
	t3, err := createCustomRelay(dbConn, relayId, cloudFunction, argument, desiredReturnCode, timeStr0, models.RelayComplete)
	if err != nil {
		t.Fatalf("TestGetReadyRelays: %+v", err)
	}
	err = testGetReadyRelay(dbConn, testTimeStr, 1, 3, []int{t0, t1})
	if err != nil {
		t.Fatalf("TestGetReadyRelays: %+v", err)
	}

	// dev0, func1, failed
	relayId = "dev0"
	t4, err := createCustomRelay(dbConn, relayId, cloudFunction, argument, desiredReturnCode, timeStr0, models.RelayFailed)
	if err != nil {
		t.Fatalf("TestGetReadyRelays: %+v", err)
	}
	err = testGetReadyRelay(dbConn, testTimeStr, 1, 3, []int{t0, t1})
	if err != nil {
		t.Fatalf("TestGetReadyRelays: %+v", err)
	}

	// dev0, func1, ready
	t5, err := createCustomRelay(dbConn, relayId, cloudFunction, argument, desiredReturnCode, timeStr0, models.RelayReady)
	if err != nil {
		t.Fatalf("TestGetReadyRelays: %+v", err)
	}
	err = testGetReadyRelay(dbConn, testTimeStr, 1, 3, []int{t0, t1})
	if err != nil {
		t.Fatalf("TestGetReadyRelays: %+v", err)
	}
	err = testGetReadyRelay(dbConn, testTimeStr, t1, 3, []int{t1, t5})
	if err != nil {
		t.Fatalf("TestGetReadyRelays: %+v", err)
	}

	// dev2, func1, ready
	relayId = "dev2"
	t6, err := createCustomRelay(dbConn, relayId, cloudFunction, argument, desiredReturnCode, timeStr0, models.RelayReady)
	if err != nil {
		t.Fatalf("TestGetReadyRelays: %+v", err)
	}
	err = testGetReadyRelay(dbConn, testTimeStr, t0, 10, []int{t0, t1, t6})
	if err != nil {
		t.Fatalf("TestGetReadyRelays: %+v", err)
	}

	// dev2, func2, ready
	cloudFunction = "func2"
	t7, err := createCustomRelay(dbConn, relayId, cloudFunction, argument, desiredReturnCode, timeStr0, models.RelayReady)
	if err != nil {
		t.Fatalf("TestGetReadyRelays: %+v", err)
	}
	err = testGetReadyRelay(dbConn, testTimeStr, t0, 10, []int{t0, t1, t6})
	if err != nil {
		t.Fatalf("TestGetReadyRelays: %+v", err)
	}

	// dev0, func2, ready, in the future
	relayId = "dev3"
	timeStr1 := "2025-05-14 20:17:32.897647+00:00"
	t8, err := createCustomRelay(dbConn, relayId, cloudFunction, argument, desiredReturnCode, timeStr1, models.RelayReady)

	if err != nil {
		t.Fatalf("TestGetReadyRelays: %+v", err)
	}
	err = testGetReadyRelay(dbConn, testTimeStr, t0, 10, []int{t0, t1, t6})
	if err != nil {
		t.Fatalf("TestGetReadyRelays: %+v", err)
	}
	t.Log(t0, t1, t2, t3, t4, t5, t6, t7, t8)
}
