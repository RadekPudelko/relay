package test

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
    "net/http"
    "strings"

	"relay/internal/client"
	"relay/internal/models"
	"relay/internal/server"
	"relay/internal/particle"
)

type TestRelay struct {
	Id       int
	DeviceId string
	DRC      int
	Status   models.RelayStatus
}

func generateRelay(nDevices int) (string, int, models.RelayStatus) {
	devNum := rand.Intn(nDevices)
	deviceId := fmt.Sprintf("dev_%d", devNum)
	drc := rand.Intn(3) + 1
	if drc == 3 {
		return deviceId, drc, models.RelayComplete
	} else {
		return deviceId, drc, models.RelayFailed
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

	// db, err := SetupMemoryDB()
	db, err := SetupFileDB("test.db3")
	if err != nil {
		t.Fatalf("TestCancellations: %+v", err)
	}
	// defer db.Close()

    particle := particle.NewMock()
	go server.BackgroundTask(config, db, particle)
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

	// Expect an error here for non existant relay
	relay, err := client.GetRelay(1)
	if err == nil {
		t.Fatalf("TestIntegration: want an error for GetRelay non existant relay got=%+v", relay)
	}
    if !strings.Contains(err.Error(), fmt.Sprintf("status code=%d", http.StatusBadRequest)) {
		t.Fatalf("TestIntegration: want %d for GetRelay on non existant relay got=%+v", http.StatusBadRequest, err)
    }

    err = client.CancelRelay(1)
    if err == nil {
		t.Fatalf("TestIntegration: want an error for CancelRelay on non existant relay got=%+v", relay)
    }
    if !strings.Contains(err.Error(), fmt.Sprintf("status code=%d", http.StatusUnprocessableEntity)) {
		t.Fatalf("TestIntegration: want %d for CancelRelay on non existant relay got %+v", http.StatusUnprocessableEntity, err)
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
	}

	// TODO: add extra routine to spam the service with gets
	for i := 0; i < nRelays; i++ {
		for {
			relay, err := client.GetRelay(testRelays[i].Id)
			if err != nil {
				t.Logf("TestIntegration: expected an error for non existant relay got %+v\n", relay)
			} else if relay.Status == models.RelayReady {
				time.Sleep(100 * time.Millisecond)
				continue
			} else if relay.Status != testRelays[i].Status {
				t.Fatalf("TestIntegration: relay status mismatch, want=%d, got=%d, relay=%+v\n", int(testRelays[i].Status), int(relay.Status), relay)
			} else {
				err = AssertRelay(relay, testRelays[i].DeviceId, cloudFunction, argument, &testRelays[i].DRC, relay.Status, scheduledTime, relay.Tries)
				if err != nil {
					t.Fatalf("TestIntegration: %+v", err)
				}
                if relay.Tries > config.MaxRetries {
                    t.Fatalf("TestIntegration: tries=%d exceeds max tries=%d\n", relay.Tries, config.MaxRetries)
                }
                break
			}
		}
	}
}
