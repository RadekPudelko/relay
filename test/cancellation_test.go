package test

import (
	"github.com/RadekPudelko/relay/pkg/models"
	"github.com/RadekPudelko/relay/internal/server"
	"testing"
	"time"
)

func TestCancellations(t *testing.T) {
	db, err := SetupMemoryDB()
	// db, err := SetupFileDB("test.db3")
	if err != nil {
		t.Fatalf("TestCancellations: %+v", err)
	}
	defer db.Close()

	id, err := models.InsertCancellation(db, 1)
	if err == nil {
		t.Fatalf("TestCancellations: expected InsertCancellation on nonexistant relay to fail, created id=%d\n", id)
	}

	relayId, err := AssertCreateRelay(db, "dev0", "", "", nil, time.Now().UTC())
	if err != nil {
		t.Fatalf("TestCancellations: %+v", err)
	}
	id, err = models.InsertCancellation(db, relayId)
	if err != nil {
		t.Fatalf("TestCancellations: %+v", err)
	}

	err = server.ProcessCancellations(db)
	if err != nil {
		t.Fatalf("TestCancellations: %+v", err)
	}

	relay, err := models.SelectRelay(db, relayId)
	if err != nil {
		t.Fatalf("TestCancellations: %+v", err)
	}
	if relay.Status != models.RelayCancelled {
		t.Fatalf("TestCancellations: relay %d status, want=%d, got=%d", relayId, models.RelayCancelled, relay.Status)
	}

	cancellations, err := models.SelectCancellations(db, 100)
	if err != nil {
		t.Fatalf("TestCancellations: %+v", err)
	}
	if len(cancellations) != 0 {
		t.Fatalf("TestCancellations: there should be no cancellations left, found %+v", cancellations)
	}
}
