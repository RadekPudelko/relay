package test

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	"relay/internal/database"
	"relay/internal/models"
	"relay/internal/server"
)

const layout = "2006-01-02 15:04:05.999999-07:00"

func SetupMemoryDB() (*sql.DB, error) {
	db, err := database.Setup(":memory:", false)
	if err != nil {
		return nil, fmt.Errorf("SetupMemoryDB: %w", err)
	}
	return db, nil
}

// TODO: Cleanup WAL files
func SetupFileDB(path string) (*sql.DB, error) {
	CleanupTestDB(path)
	db, err := database.Setup(path, true)
	if err != nil {
		return nil, fmt.Errorf("SetupFileDB: %w", err)
	}
	return db, nil
}

func CleanupTestDB(path string) {
	if _, err := os.Stat(path); err == nil {
		os.Remove(path)
	}
}

func AssertCreateRelay(db *sql.DB,
	deviceId string,
	cloudFunction string,
	argument string,
	desiredReturnCode sql.NullInt64,
	scheduledTime time.Time,
) (int, error) {
	id, err := server.CreateRelay(db, deviceId, cloudFunction, argument, desiredReturnCode, scheduledTime)
	if err != nil {
		return 0, fmt.Errorf("AssertCreateRelay: %w", err)
	}

	relay, err := models.SelectRelay(db, id)
	if err != nil {
		return 0, fmt.Errorf("AssertCreateRelay: %w", err)
	}

	if relay.Device.DeviceId != deviceId {
		return 0, fmt.Errorf("AssertCreateRelay: SelectRelay on relay %d got relayId %s, expected %s", id, relay.Device.DeviceId, deviceId)
	}
	if relay.CloudFunction != cloudFunction {
		return 0, fmt.Errorf("AssertCreateRelay: SelectRelay on relay %d got cloudFunction %s, expected %s", id, relay.CloudFunction, cloudFunction)
	}
	if relay.Argument != argument {
		return 0, fmt.Errorf("AssertCreateRelay: SelectRelay on relay %d got argument %s, expected %s", id, relay.Argument, argument)
	}
	if relay.DesiredReturnCode != desiredReturnCode {
		return 0, fmt.Errorf("AssertCreateRelay: SelectRelay on relay %d got desiredReturnCode %+v, expected %+v", id, relay.DesiredReturnCode, desiredReturnCode)
	}
	if relay.ScheduledTime != scheduledTime {
		return 0, fmt.Errorf("AssertCreateRelay: SelectRelay on relay %d got scheduledTime %s, expected %s", id, relay.ScheduledTime, scheduledTime)
	}
	return id, nil
}

func AssertUpdateRelay(db *sql.DB, relayId int, scheduledTime time.Time, status models.RelayStatus, tries int) error {
	err := models.UpdateRelay(db, relayId, scheduledTime, status, tries)
	if err != nil {
		return fmt.Errorf("AssertUpdateRelay: %w", err)
	}
	relay, err := models.SelectRelay(db, relayId)
	if err != nil {
		return fmt.Errorf("AssertUpdateRelay: %w", err)
	}
	if relay.ScheduledTime != scheduledTime {
		return fmt.Errorf("AssertUpdateRelay: scheduleTime want=%s, got=%s", scheduledTime, relay.ScheduledTime)
	}
	if relay.ScheduledTime != scheduledTime {
		return fmt.Errorf("AssertUpdateRelay: status want=%d, got=%d", status, relay.Status)
	}
	if relay.Tries != tries {
		return fmt.Errorf("AssertUpdateRelay: tries want=%d, got=%d", tries, relay.Tries)
	}
	return nil
}

func AssertCreateAndUpdateRelay(db *sql.DB, deviceId string, cloudFunction, argument string, desiredReturnCode sql.NullInt64, scheduledTime time.Time, status models.RelayStatus, tries int) (int, error) {
	relayId, err := AssertCreateRelay(db, deviceId, cloudFunction, argument, desiredReturnCode, scheduledTime)
	if err != nil {
		return 0, fmt.Errorf("AssertCreateAndUpdateRelay: %w", err)
	}
	err = AssertUpdateRelay(db, relayId, scheduledTime, status, tries)
	if err != nil {
		return 0, fmt.Errorf("AssertCreateAndUpdateRelay: %w", err)
	}
	return relayId, nil
}

func AssertGetReadyRelays(db *sql.DB, scheduledTime time.Time, startId, limit int, expectedRelayIds []int) error {
	relays, err := server.GetReadyRelays(db, startId, limit, scheduledTime)
	if err != nil {
		return fmt.Errorf("AssertGetReadyRelays: %w", err)
	}
	if !SliceCompare(relays, expectedRelayIds) {
		return fmt.Errorf("TestGetReadyRelays: mismatch relays, expected %+v, got %+v", expectedRelayIds, relays)
	}
	return nil
}

func SliceCompare(a, b []int) bool {
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

// func runTestServer(config server.Config) error {
// 	particle := particle.NewMock()
//
// 	// TODO: Load from .env
// 	// testDBPath := ":memory:"
// 	// testDBPath := "test/test.db3"
// 	testDBPath := "test.db3"
// 	// _, err := os.Stat(testDBPath)
// 	// fmt.Printf("%+v\n", err)
// 	// if err != nil && !os.IsNotExist(err) {
// 	//     return fmt.Errorf("run: %w", err)
// 	// } else if err != nil {
// 	err := os.Remove(testDBPath)
// 	if err != nil {
// 		return fmt.Errorf("run: %w", err)
// 	}
// 	// }
// 	// return fmt.Errorf("asdf")
// 	// testDBPath += "?cache=shared?_busy_timeout=5000"
// 	testDBPath += "?cache=shared"
//
// 	db, err := database.Connect(testDBPath)
// 	if err != nil {
// 		return fmt.Errorf("run: %w", err)
// 	}
// 	defer db.Close()
//
// 	// https://phiresky.github.io/blog/2020/sqlite-performance-tuning/
// 	_, err = db.Exec("PRAGMA journal_mode=WAL;")
// 	if err != nil {
// 		return fmt.Errorf("run: %w", err)
// 	}
//
// 	// Confirm that WAL mode is enabled
// 	var mode string
// 	err = db.QueryRow("PRAGMA journal_mode;").Scan(&mode)
// 	if err != nil {
// 		return fmt.Errorf("run: %w", err)
// 	}
// 	if mode != "wal" {
// 		return fmt.Errorf("run: expected wal mode")
// 	}
//
// 	err = database.CreateTables(db)
// 	if err != nil {
// 		return fmt.Errorf("run: %w", err)
// 	}
//
// 	err = server.Run(config, db, particle)
// 	if err != nil {
// 		return fmt.Errorf("run: %w", err)
// 	}
//
// 	return nil
// }
