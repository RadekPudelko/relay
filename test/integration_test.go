package test

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"relay/internal/client"
	"relay/internal/models"
	"relay/internal/particle"
	"relay/internal/server"
	"relay/internal/config"
)

type TestRelay struct {
	Id       int
	DeviceId string
	DRC      int
	Status   models.RelayStatus
	Cancel   bool
}

func generateRelay(nDevices int) (string, int, models.RelayStatus) {
	devNum := rand.Intn(nDevices)
	deviceId := fmt.Sprintf("dev_%d", devNum)
	drc := rand.Intn(3) + 1 // 1-3 correspond to MockParticle return status options DRC != status
	if drc == 3 {           // Success
		return deviceId, drc, models.RelayComplete
	} else {
		return deviceId, drc, models.RelayFailed
	}
}

func TestIntegration(t *testing.T) {
	t.Log("TestIntegration")

    myConfig := config.GetDefaultConfig()
    myConfig.Settings.PingRetrySeconds = 15
    myConfig.Settings.CFRetrySeconds = 10

	// db, err := SetupMemoryDB()
	db, err := SetupFileDB("test.db3")
	if err != nil {
		t.Fatalf("TestCancellations: %+v", err)
	}
	// defer db.Close()

	particle := particle.NewMock()
	go server.BackgroundTask(&myConfig, db, particle)
	go func() {
		if err := server.Run(db, "localhost", "8080"); err != nil {
			// TODO: Fix this warning
			t.Fatalf("TestCancellations: Could not start server: %s\n", err)
		}
	}()

	time.Sleep(100 * time.Millisecond)
	client := client.NewClient(8080)
	err = client.Ping()
	if err != nil {
		t.Fatalf("TestIntegration: %+v", err)
	}

	cloudFunction := "func0"
	argument := ""
	var scheduledTime *time.Time = nil

	nRelays := 1000
	nDevices := 20
	testRelays := make([]TestRelay, nRelays)

	// TODO: use goroutines to hit in parallel?
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
		testRelays[i].Cancel = rand.Intn(10) == 0
		if testRelays[i].Cancel {
			err = client.CancelRelay(id)
			if err != nil {
				t.Logf("TestIntegration: %+v", err)
			}
		}
	}

	// TODO: add extra routine to spam the service with gets
	// TODO: Move onto next relay if it is still in ready state
	for i := 0; i < nRelays; i++ {
		for {
			relay, err := client.GetRelay(testRelays[i].Id)
			if err != nil {
				t.Logf("TestIntegration: expected an error for non existant relay got %+v\n", relay)
			} else if relay.Status == models.RelayReady {
				time.Sleep(100 * time.Millisecond)
				continue
			} else if relay.Status == testRelays[i].Status ||
				testRelays[i].Cancel && relay.Status == models.RelayCancelled {
				err = AssertRelay(relay, testRelays[i].DeviceId, cloudFunction, argument, &testRelays[i].DRC, relay.Status, scheduledTime, relay.Tries)
				if err != nil {
					t.Fatalf("TestIntegration: %+v", err)
				}
				if relay.Tries > myConfig.Settings.MaxRetries {
					t.Fatalf("TestIntegration: tries=%d exceeds max tries=%d\n", relay.Tries, myConfig.Settings.MaxRetries)
				}
				break
			} else {
				t.Fatalf("TestIntegration: relay status mismatch, want=%d, got=%d, relay=%+v, testRelay=%+v\n", int(testRelays[i].Status), int(relay.Status), relay, testRelays[i])
			}
		}
	}
}
