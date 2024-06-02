package test

import (
	"fmt"
	"testing"
	"time"
    "net/http"
    "strings"

	"relay/internal/client"
	"relay/internal/models"
	"relay/internal/server"
	"relay/internal/particle"
)

func TestClient(t *testing.T) {
	t.Log("TestClient")
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
		t.Fatalf("TestClient: %+v", err)
	}

	// Expect an error here for non existant relay
	relay, err := client.GetRelay(1)
	if err == nil {
		t.Fatalf("TestClient: want an error for GetRelay non existant relay got=%+v", relay)
	}
    if !strings.Contains(err.Error(), fmt.Sprintf("status code=%d", http.StatusBadRequest)) {
		t.Fatalf("TestClient: want %d for GetRelay on non existant relay got=%+v", http.StatusBadRequest, err)
    }

    err = client.CancelRelay(1)
    if err == nil {
		t.Fatalf("TestClient: want an error for CancelRelay on non existant relay got=%+v", relay)
    }
    if !strings.Contains(err.Error(), fmt.Sprintf("status code=%d", http.StatusUnprocessableEntity)) {
		t.Fatalf("TestClient: want %d for CancelRelay on non existant relay got %+v", http.StatusUnprocessableEntity, err)
    }

    deviceId := "device0"
	cloudFunction := "func0"
	argument := ""
    var drc *int = nil
	var scheduledTime *time.Time = nil

    id, err := client.CreateRelay(deviceId, cloudFunction, argument, drc, scheduledTime)
    if err != nil {
        t.Fatalf("TestClient: %+v", err)
    }

	relay, err = client.GetRelay(id)
	if err != nil {
        t.Fatalf("TestClient: %+v", err)
	}
    err = AssertRelay(relay, deviceId, cloudFunction, argument, drc, models.RelayReady, scheduledTime, 0)
	if err != nil {
        t.Fatalf("TestClient: %+v", err)
	}

    err = client.CancelRelay(id)
    if err != nil {
        t.Fatalf("TestClient: %+v", err)
    }

	go server.BackgroundTask(config, db, particle)
    time.Sleep(100 * time.Millisecond)

	relay, err = client.GetRelay(id)
	if err != nil {
        t.Fatalf("TestClient: %+v", err)
	}
    err = AssertRelay(relay, deviceId, cloudFunction, argument, drc, models.RelayCancelled, scheduledTime, 0)
	if err != nil {
        t.Fatalf("TestClient: %+v", err)
	}
}
