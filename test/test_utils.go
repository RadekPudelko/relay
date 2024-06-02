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
	desiredReturnCode *int,
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
    err = AssertRelay(relay, deviceId, cloudFunction, argument, desiredReturnCode, models.RelayReady, &scheduledTime, 0)
    if err != nil {
        return 0, fmt.Errorf("AssertCreateRelay: %w", err)
    }
	return id, nil
}

func AssertRelay(
	relay *models.Relay,
	deviceId string,
	cloudFunction string,
	argument string,
	desiredReturnCode *int,
    status models.RelayStatus,
	scheduledTime *time.Time,
    tries int,
) error {
	if relay.Device.DeviceId != deviceId {
		return fmt.Errorf("AssertRelay: deviceId, want=%s, got=%s", deviceId, relay.Device.DeviceId)
	}
	if relay.CloudFunction != cloudFunction {
		return fmt.Errorf("AssertRelay: cloudFunction, want=%s, got=%s", cloudFunction, relay.CloudFunction)
	}
	if relay.Argument != argument {
		return fmt.Errorf("AssertRelay: argument, want=%s, got=%s", argument, relay.Argument)
	}
	if desiredReturnCode != nil {
        if relay.DesiredReturnCode == nil {
			return fmt.Errorf("AssertRelay: desired return code: got nil")
        } else if *relay.DesiredReturnCode != *desiredReturnCode {
			return fmt.Errorf("AssertRelay: desired return code: want=%d, got=%d", *desiredReturnCode, *relay.DesiredReturnCode)
		}
	}
	if relay.Status != status {
		return fmt.Errorf("AssertRelay: status want=%d, got=%d", status, relay.Status)
	}
	if scheduledTime != nil && relay.ScheduledTime != *scheduledTime {
		return fmt.Errorf("AssertRelay: scheduled time: want=%s, got=%s", scheduledTime, relay.ScheduledTime)
	}
	if relay.Tries != tries {
		return fmt.Errorf("AssertRelay: tries want=%d, got=%d", tries, relay.Tries)
	}
	return nil
}

func AssertUpdateRelay(db *sql.DB, relayId int, deviceId string, cloudFunction string, argument string, desiredReturnCode *int, scheduledTime time.Time, status models.RelayStatus, tries int) error {
	err := models.UpdateRelay(db, relayId, scheduledTime, status, tries)
	if err != nil {
		return fmt.Errorf("AssertUpdateRelay: %w", err)
	}
	relay, err := models.SelectRelay(db, relayId)
	if err != nil {
		return fmt.Errorf("AssertUpdateRelay: %w", err)
	}
    err = AssertRelay(relay, deviceId, cloudFunction, argument, desiredReturnCode, status, &scheduledTime, tries)
	return nil
}

func AssertCreateAndUpdateRelay(db *sql.DB, deviceId string, cloudFunction, argument string, desiredReturnCode *int, scheduledTime time.Time, status models.RelayStatus, tries int) (int, error) {
	relayId, err := AssertCreateRelay(db, deviceId, cloudFunction, argument, desiredReturnCode, scheduledTime)
	if err != nil {
		return 0, fmt.Errorf("AssertCreateAndUpdateRelay: %w", err)
	}
	err = AssertUpdateRelay(db, relayId, deviceId, cloudFunction, argument, desiredReturnCode, scheduledTime, status, tries)
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
		return fmt.Errorf("TestGetReadyRelays: mismatch relays, want %+v, got %+v", expectedRelayIds, relays)
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
