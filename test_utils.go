package main

import (
    "fmt"

	"pcfs/db"
	"pcfs/particle"
	"pcfs/server"
)

func runTestServer() error {
	config := server.Config{
		Host:        "localhost",
		Port:        "8080",
		MaxRoutines: 2,
		TaskLimit:   10,
		MaxRetries:  3,
	}

    particle := particle.NewMock()

	// testDBPath := "test.db3"
	// err := os.Remove(testDBPath)
	// if err != nil {
	//        return fmt.Errorf("run: %w", err)
	// }
    testDBPath := ":memory:"
    testDBPath += "?cache=shared"

	dbConn, err := db.Connect(testDBPath)
	if err != nil {
        return fmt.Errorf("run: %w", err)
	}
    defer dbConn.Close()

	err = db.CreateTables(dbConn)
	if err != nil {
		return fmt.Errorf("run: %w", err)
	}

	err = Run(config, dbConn, particle)
	if err != nil {
        return fmt.Errorf("run: %w", err)
    }

	return nil
}

