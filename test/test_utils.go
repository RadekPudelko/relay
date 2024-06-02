package test

import (
	"fmt"
	"os"
	"database/sql"
    "time"

	"relay/internal/database"
	"relay/internal/server"
	"relay/internal/models"
)


func SetupMemoryDB() (*sql.DB, error) {
    db, err := database.Setup(":memory:", false)
	if err != nil {
        return nil, fmt.Errorf("SetupMemoryDB: %w", err)
	}
    return db, nil
}

func SetupFileDB(path string) (*sql.DB, error) {
    CleanupTestDB(path)
    db, err := database.Setup(path, false)
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

func AssertCreateRelay(dbConn *sql.DB,
    deviceId string,
    cloudFunction string,
    argument string,
    desiredReturnCode sql.NullInt64,
    scheduledTime time.Time,
) (int, error) {
	id, err := server.CreateRelay(dbConn, deviceId, cloudFunction, argument, desiredReturnCode, scheduledTime)
	if err != nil {
		return 0, fmt.Errorf("AssertCreateRelay: %w", err)
	}

	relay, err := models.SelectRelay(dbConn, id)
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
// 	dbConn, err := database.Connect(testDBPath)
// 	if err != nil {
// 		return fmt.Errorf("run: %w", err)
// 	}
// 	defer dbConn.Close()
//
// 	// https://phiresky.github.io/blog/2020/sqlite-performance-tuning/
// 	_, err = dbConn.Exec("PRAGMA journal_mode=WAL;")
// 	if err != nil {
// 		return fmt.Errorf("run: %w", err)
// 	}
//
// 	// Confirm that WAL mode is enabled
// 	var mode string
// 	err = dbConn.QueryRow("PRAGMA journal_mode;").Scan(&mode)
// 	if err != nil {
// 		return fmt.Errorf("run: %w", err)
// 	}
// 	if mode != "wal" {
// 		return fmt.Errorf("run: expected wal mode")
// 	}
//
// 	err = database.CreateTables(dbConn)
// 	if err != nil {
// 		return fmt.Errorf("run: %w", err)
// 	}
//
// 	err = server.Run(config, dbConn, particle)
// 	if err != nil {
// 		return fmt.Errorf("run: %w", err)
// 	}
//
// 	return nil
// }
