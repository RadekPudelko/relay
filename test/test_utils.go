package test

import (
	"fmt"
	"os"

	"relay/db"
	"relay/particle"
	"relay/server"
)

func runTestServer() error {
	config := server.Config{
		Host:        "localhost",
		Port:        "8080",
		MaxRoutines: 1,
		TaskLimit:   10,
		MaxRetries:  3,
	}

	particle := particle.NewMock()

    // TODO: Load from .env
	// testDBPath := ":memory:"
	// testDBPath := "test/test.db3"
	testDBPath := "test.db3"
    // _, err := os.Stat(testDBPath)
    // fmt.Printf("%+v\n", err)
    // if err != nil && !os.IsNotExist(err) {
    //     return fmt.Errorf("run: %w", err)
    // } else if err != nil {
        err := os.Remove(testDBPath)
        if err != nil {
            return fmt.Errorf("run: %w", err)
        }
    // }
    // return fmt.Errorf("asdf")
	testDBPath += "?cache=shared"

	dbConn, err := db.Connect(testDBPath)
	if err != nil {
		return fmt.Errorf("run: %w", err)
	}
	defer dbConn.Close()

    // https://phiresky.github.io/blog/2020/sqlite-performance-tuning/
    _, err = dbConn.Exec("PRAGMA journal_mode=WAL;")
    if err != nil {
		return fmt.Errorf("run: %w", err)
    }

    // Confirm that WAL mode is enabled
    var mode string
    err = dbConn.QueryRow("PRAGMA journal_mode;").Scan(&mode)
    if err != nil {
		return fmt.Errorf("run: %w", err)
    }
    if mode != "wal" {
		return fmt.Errorf("run: expected wal mode")
    }


	err = db.CreateTables(dbConn)
	if err != nil {
		return fmt.Errorf("run: %w", err)
	}

	err = server.Run(config, dbConn, particle)
	if err != nil {
		return fmt.Errorf("run: %w", err)
	}

	return nil
}
