package test

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"relay/client"
	"relay/db"
	"relay/server"
	// "relay/particle"
)

type TestRelay struct {
	Id       int
	DeviceId string
	DRC      int
	Status   db.RelayStatus
}

func assertRelay(
	relay *db.Relay,
	config server.Config,
	deviceId string,
	cloudFunction string,
	argument string,
	desiredReturnCode *int,
	scheduledTime *time.Time,
) error {
	if relay.Device.DeviceId != deviceId {
		return fmt.Errorf("assertRelay: deviceId, expected=%s, got=%s", deviceId, relay.Device.DeviceId)
	}
	if relay.CloudFunction != cloudFunction {
		return fmt.Errorf("assertRelay: cloudFunction, expected=%s, got=%s", cloudFunction, relay.CloudFunction)
	}
	if relay.Argument != argument {
		return fmt.Errorf("assertRelay: argument, expected=%s, got=%s", argument, relay.Argument)
	}
	if desiredReturnCode != nil {
		if !relay.DesiredReturnCode.Valid {
			return fmt.Errorf("assertRelay: desired return code: got invalid")
		} else if int(relay.DesiredReturnCode.Int64) != *desiredReturnCode {
			return fmt.Errorf("assertRelay: desired return code: expected=%d, got=%d", *desiredReturnCode, relay.DesiredReturnCode.Int64)
		}
	}
	if scheduledTime != nil && relay.ScheduledTime != *scheduledTime {
		return fmt.Errorf("assertRelay: scheduled time: expected=%s, got=%s", scheduledTime, relay.ScheduledTime)
	}
	if relay.Tries > config.MaxRetries {
		return fmt.Errorf("assertRelay: tries=%d exceeds max tries=%d\n", relay.Tries, config.MaxRetries)
	}
	return nil
}

func generateRelay(nDevices int) (string, int, db.RelayStatus) {
	devNum := rand.Intn(nDevices)
	deviceId := fmt.Sprintf("dev_%d", devNum)
	drc := rand.Intn(3) + 1
	if drc == 3 {
		return deviceId, drc, db.RelayComplete
	} else {
		return deviceId, drc, db.RelayFailed
	}
}

func TestIntegration(t *testing.T) {
	t.Log("TestIntegration")
	config := server.Config{
		Host:              "localhost",
		Port:              "8080",
		MaxRoutines:       3,
		RelayLimit:        10,
		PingRetryDuration: 15 * time.Second,
		CFRetryDuration:   10 * time.Second,
		MaxRetries:        3,
	}
	go func() {
		err := runTestServer(config)
		if err != nil {
			// TODO: Fix this warning
			t.Fatalf("TestIntegration: %+v", err)
		}
	}()

	time.Sleep(50 * time.Millisecond)
	client := client.NewClient(8080)
	err := client.Ping()
	if err != nil {
		t.Fatalf("TestIntegration: %+v", err)
	}

	// Expect an error here for non existant relay
	relay, err := client.GetRelay(1)
	if err == nil {
		t.Fatalf("TestIntegration: expected an error for non existant relay got %+v", relay)
	}

	cloudFunction := "func0"
	argument := ""
	var scheduledTime *time.Time = nil

	nRelays := 1000
	nDevices := 20
	testRelays := make([]TestRelay, nRelays)

	// TODO: use goroutines to hit multiple requests as faster
	for i := 0; i < nRelays; i++ {
		deviceId, drc, status := generateRelay(nDevices)
		id, err := client.CreateRelay(deviceId, cloudFunction, argument, &drc, scheduledTime)
		if err != nil {
			t.Fatalf("TestIntegration: %+v", err)
		}
		testRelays[i].Id = id
		testRelays[i].DeviceId = deviceId
		testRelays[i].DRC = drc
		testRelays[i].Status = status
	}

	// TODO: add extra routine to spam the service with gets
	for i := 0; i < nRelays; i++ {
		for {
			relay, err := client.GetRelay(testRelays[i].Id)
			if err != nil {
				t.Logf("TestIntegration: expected an error for non existant relay got %+v\n", relay)
			} else if relay.Status == db.RelayReady {
				time.Sleep(100 * time.Millisecond)
				continue
			} else if relay.Status != testRelays[i].Status {
				t.Fatalf("TestIntegration: relay status mismatch, want=%d, got=%d, relay=%+v\n", int(testRelays[i].Status), int(relay.Status), relay)
			} else {
				err = assertRelay(relay, config, testRelays[i].DeviceId, cloudFunction, argument, &testRelays[i].DRC, scheduledTime)
				if err != nil {
					t.Fatalf("TestIntegration: %+v", err)
				}
				break
			}
		}
	}
}
