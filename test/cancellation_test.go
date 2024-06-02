package test

import (
	"database/sql"
	"testing"
	"time"
	"relay/internal/models"
	"relay/internal/server"
)

func TestCancellations(t *testing.T) {
    db, err := SetupFileDB("test.db3")
	if err != nil {
		t.Fatalf("TestCancellations: %+v", err)
	}
    defer db.Close()

    id, err := models.InsertCancellation(db, 1)
    if err == nil {
        t.Fatalf("TestCancellations: expected InsertCancellation on nonexistant relay to fail, created id=%d\n", id)
    }

    relayId, err:= AssertCreateRelay(db, "dev0", "", "", sql.NullInt64{Int64: 0, Valid: false}, time.Now().UTC())
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







    // TODO: Close the server at the end of the test?
    // go func() {
    //     if err := server.Run(db, "localhost", "8080"); err != nil {
    //         t.Fatalf("TestCancellations: Could not start server: %s\n", err)
    //     }
    // }()
    // time.Sleep(100 * time.Millisecond)

	// client := client.NewClient(8080)
	// err = client.Ping()
	// if err != nil {
	//        t.Fatalf("TestCancellations: failed to ping the server")
	// }

    // req, err := http.NewRequest("DELETE", "localhost:8080/api/relays/1", nil)
    // if err != nil {
    //     t.Fatalf("TestCancellations: %+v", err)
    // }
    //
    // if req.Response.StatusCode != http.StatusUnprocessableEntity {
    //     t.Errorf("TestCancellations: handler returned wrong status code for nonexistant relay: got %v want %v", req.Response.StatusCode, http.StatusUnprocessableEntity)
    // }
    // t.Logf("%+v\n", err)
    // t.Logf("%+v\n", req)
	// server := httptest.NewServer(server.HandleCancelRelay(db))
	// server := httptest.NewServer(server.HandleCancelRelay(db))

    // cancelHandler.ServeHTTP(rr, req)
    // if status := rr.Code; status != http.StatusUnprocessableEntity {
    // }
	// defer server.Close()

	// req, err := http.NewRequest("Delete", server.URL +
	// if err != nil {
	// 	return fmt.Errorf("CancelRelay: http.NewRequest: %w", err)
	// }
	//
	// req.Header.Set("Content-Type", "application/json")
	//
	// client := &http.Client{}
	// resp, err := client.Do(req)
	// if err != nil {
	// 	return fmt.Errorf("CancelRelay: client.Do: %w", err)
	// }
	// defer resp.Body.Close()
	//
	// body, err := io.ReadAll(resp.Body)
	// if err != nil {
	// 	return fmt.Errorf("CancelRelay: io.ReadAll: %w", err)
	// }
	//
	// // Check the response status code
	// if resp.StatusCode != http.StatusAccepted {
	// 	return fmt.Errorf("CancelRelay: request error, status code=%d, body=%s", resp.StatusCode, body)
	// }
	//
	// resp, err := http.Get(server.URL)
	// if err != nil {
	// 	t.Fatalf("error making request to server. Err: %v", err)
	// }
	// defer resp.Body.Close()
	// // Assertions
	// if resp.StatusCode != http.StatusOK {
	// 	t.Errorf("expected status OK; got %v", resp.Status)
	// }
	// expected := "{\"message\":\"Hello World\"}"
	// body, err := io.ReadAll(resp.Body)
	// if err != nil {
	// 	t.Fatalf("error reading response body. Err: %v", err)
	// }
	// if expected != string(body) {
	// 	t.Errorf("expected response body to be %v; got %v", expected, string(body))
	// }
	// defer dbConn.Close()
	//
	// err = database.CreateTables(dbConn)
	// if err != nil {
	// 	t.Fatalf("TestCancellations: %+v", err)
	// }
	//
	// relayId := "devid0"
	// cloudFunction := "func1"
	// argument := ""
	// desiredReturnCode := sql.NullInt64{Int64: 0, Valid: false}
	// scheduledTime0 := time.Now().UTC()
	//
	//
	// tid, err := testCreateRelay(dbConn, relayId, cloudFunction, argument, desiredReturnCode, scheduledTime0)
	// if err != nil {
	// 	t.Fatalf("TestCancellations: %+v", err)
	// }
	// if tid != 1 {
	// 	t.Fatalf("TestCancellations: expected to create relay id 1, got %d", tid)
	// }
	//
	// relayId = "devid1"
	// tid, err = testCreateRelay(dbConn, relayId, cloudFunction, argument, desiredReturnCode, scheduledTime0)
	// if err != nil {
	// 	t.Fatalf("TestCancellations: %+v", err)
	// }
	// if tid != 2 {
	// 	t.Fatalf("TestCancellations: expected to create relay id 2, got %d", tid)
	// }
	//
	// desiredReturnCode = sql.NullInt64{Int64: 0, Valid: true}
	// tid, err = testCreateRelay(dbConn, relayId, cloudFunction, argument, desiredReturnCode, scheduledTime0)
	// if err != nil {
	// 	t.Fatalf("TestCancellations: %+v", err)
	// }
	// if tid != 3 {
	// 	t.Fatalf("TestCancellations: expected to create relay id 3, got %d", tid)
	// }
}
