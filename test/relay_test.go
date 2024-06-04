package test

import (
	"testing"
	"time"

	"github.com/RadekPudelko/relay/pkg/models"
)

func TestCreateRelay(t *testing.T) {
	// db, err := SetupMemoryDB()
	db, err := SetupFileDB("test.db3")
	if err != nil {
		t.Fatalf("TestCancellations: %+v", err)
	}
	defer db.Close()

	relayId := "devid0"
	cloudFunction := "func1"
	argument := ""
	var desiredReturnCode *int = nil
	scheduledTime0 := time.Now().UTC()

	tid, err := AssertCreateRelay(db, relayId, cloudFunction, argument, desiredReturnCode, scheduledTime0)
	if err != nil {
		t.Fatalf("TestCreateRelays: %+v", err)
	}
	if tid != 1 {
		t.Fatalf("TestCreateRelays: expected to create relay id 1, got %d", tid)
	}

	relayId = "devid1"
	tid, err = AssertCreateRelay(db, relayId, cloudFunction, argument, desiredReturnCode, scheduledTime0)
	if err != nil {
		t.Fatalf("TestCreateRelays: %+v", err)
	}
	if tid != 2 {
		t.Fatalf("TestCreateRelays: expected to create relay id 2, got %d", tid)
	}

	code := 0
	desiredReturnCode = &code
	tid, err = AssertCreateRelay(db, relayId, cloudFunction, argument, desiredReturnCode, scheduledTime0)
	if err != nil {
		t.Fatalf("TestCreateRelays: %+v", err)
	}
	if tid != 3 {
		t.Fatalf("TestCreateRelays: expected to create relay id 3, got %d", tid)
	}
}

func TestGetReadyRelays(t *testing.T) {
	db, err := SetupMemoryDB()
	// db, err := SetupFileDB("test.db3")
	if err != nil {
		t.Fatalf("TestCancellations: %+v", err)
	}
	defer db.Close()

	relayId := "dev0"
	cloudFunction := "func0"
	argument := ""
	var desiredReturnCode *int = nil

	testTimeStr := "2024-05-15 20:17:32.897647+00:00" // 1 day after scheduled relays
	testTime, err := time.Parse(layout, testTimeStr)
	if err != nil {
		t.Fatalf("TestGetReadyRelays: time.Parse on %s", testTimeStr)
	}

	testTime = testTime.UTC()
	timeStr0 := "2024-05-14 20:17:32.897647+00:00"
	scheduledTime0, err := time.Parse(layout, timeStr0)
	if err != nil {
		t.Fatalf("TestGetReadyRelays: time.Parse on %s", timeStr0)
	}
	scheduledTime0 = scheduledTime0.UTC()

	// No relays
	err = AssertGetReadyRelays(db, testTime, 0, 10, []int{})
	if err != nil {
		t.Fatalf("TestGetReadyRelays: %+v", err)
	}

	// dev0
	t0, err := AssertCreateAndUpdateRelay(db, relayId, cloudFunction, argument, desiredReturnCode, scheduledTime0, models.RelayReady, 1)
	if err != nil {
		t.Fatalf("TestGetReadyRelays: %+v", err)
	}
	err = AssertGetReadyRelays(db, testTime, 1, 1, []int{t0})
	if err != nil {
		t.Fatalf("TestGetReadyRelays: %+v", err)
	}

	// dev1
	relayId = "dev1"
	t1, err := AssertCreateAndUpdateRelay(db, relayId, cloudFunction, argument, desiredReturnCode, scheduledTime0, models.RelayReady, 1)
	if err != nil {
		t.Fatalf("TestGetReadyRelays: %+v", err)
	}
	err = AssertGetReadyRelays(db, testTime, 1, 1, []int{t0})
	if err != nil {
		t.Fatalf("TestGetReadyRelays: %+v", err)
	}
	err = AssertGetReadyRelays(db, testTime, 1, 2, []int{t0, t1})
	if err != nil {
		t.Fatalf("TestGetReadyRelays: %+v", err)
	}
	// err = AssertGetReadyRelays(db, testTime, 2, 3, []int{t1, t0}) // TODO: support wrap
	err = AssertGetReadyRelays(db, testTime, 2, 3, []int{t1})
	if err != nil {
		t.Fatalf("TestGetReadyRelays: %+v", err)
	}

	// dev1, func1
	cloudFunction = "func1"
	t2, err := AssertCreateAndUpdateRelay(db, relayId, cloudFunction, argument, desiredReturnCode, scheduledTime0, models.RelayReady, 1)
	if err != nil {
		t.Fatalf("TestGetReadyRelays: %+v", err)
	}
	err = AssertGetReadyRelays(db, testTime, 1, 3, []int{t0, t1})
	if err != nil {
		t.Fatalf("TestGetReadyRelays: %+v", err)
	}

	// dev1, func1, complete
	t3, err := AssertCreateAndUpdateRelay(db, relayId, cloudFunction, argument, desiredReturnCode, scheduledTime0, models.RelayComplete, 1)
	if err != nil {
		t.Fatalf("TestGetReadyRelays: %+v", err)
	}
	err = AssertGetReadyRelays(db, testTime, 1, 3, []int{t0, t1})
	if err != nil {
		t.Fatalf("TestGetReadyRelays: %+v", err)
	}

	// dev0, func1, failed
	relayId = "dev0"
	t4, err := AssertCreateAndUpdateRelay(db, relayId, cloudFunction, argument, desiredReturnCode, scheduledTime0, models.RelayFailed, 1)
	if err != nil {
		t.Fatalf("TestGetReadyRelays: %+v", err)
	}
	err = AssertGetReadyRelays(db, testTime, 1, 3, []int{t0, t1})
	if err != nil {
		t.Fatalf("TestGetReadyRelays: %+v", err)
	}

	// dev0, func1, ready
	t5, err := AssertCreateAndUpdateRelay(db, relayId, cloudFunction, argument, desiredReturnCode, scheduledTime0, models.RelayReady, 1)
	if err != nil {
		t.Fatalf("TestGetReadyRelays: %+v", err)
	}
	err = AssertGetReadyRelays(db, testTime, 1, 3, []int{t0, t1})
	if err != nil {
		t.Fatalf("TestGetReadyRelays: %+v", err)
	}
	err = AssertGetReadyRelays(db, testTime, t1, 3, []int{t1, t5})
	if err != nil {
		t.Fatalf("TestGetReadyRelays: %+v", err)
	}

	// dev2, func1, ready
	relayId = "dev2"
	t6, err := AssertCreateAndUpdateRelay(db, relayId, cloudFunction, argument, desiredReturnCode, scheduledTime0, models.RelayReady, 1)
	if err != nil {
		t.Fatalf("TestGetReadyRelays: %+v", err)
	}
	err = AssertGetReadyRelays(db, testTime, t0, 10, []int{t0, t1, t6})
	if err != nil {
		t.Fatalf("TestGetReadyRelays: %+v", err)
	}

	// dev2, func2, ready
	cloudFunction = "func2"
	t7, err := AssertCreateAndUpdateRelay(db, relayId, cloudFunction, argument, desiredReturnCode, scheduledTime0, models.RelayReady, 1)
	if err != nil {
		t.Fatalf("TestGetReadyRelays: %+v", err)
	}
	err = AssertGetReadyRelays(db, testTime, t0, 10, []int{t0, t1, t6})
	if err != nil {
		t.Fatalf("TestGetReadyRelays: %+v", err)
	}

	// dev0, func2, ready, in the future
	relayId = "dev3"
	timeStr1 := "2025-05-14 20:17:32.897647+00:00"
	scheduledTime1, err := time.Parse(layout, timeStr1)
	if err != nil {
		t.Fatalf("TestGetReadyRelays: time.Parse on %s", timeStr0)
	}
	scheduledTime1 = scheduledTime1.UTC()
	t8, err := AssertCreateAndUpdateRelay(db, relayId, cloudFunction, argument, desiredReturnCode, scheduledTime1, models.RelayReady, 1)

	if err != nil {
		t.Fatalf("TestGetReadyRelays: %+v", err)
	}
	err = AssertGetReadyRelays(db, testTime, t0, 10, []int{t0, t1, t6})
	if err != nil {
		t.Fatalf("TestGetReadyRelays: %+v", err)
	}
	t.Log(t0, t1, t2, t3, t4, t5, t6, t7, t8)
}
