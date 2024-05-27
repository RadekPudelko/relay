package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"pcfs/db"
	"pcfs/particle"
)

var taskLimit = 10

func SetupDB(path string) (*sql.DB, error) {
	var err error
    dbConn, err := db.Connect(path)
	if err != nil {
		return nil, fmt.Errorf("SetupDB: %w", err)
	}

	err = db.CreateTables(dbConn)
	if err != nil {
		dbConn.Close()
		return nil, fmt.Errorf("SetupDB: %w", err)
	}
	return dbConn, nil
}

func run(dbConn *sql.DB, particle1 particle.ParticleProvider) (error) {
    // TODO: add background task to the server
	go backgroundTask(dbConn, particle1)

    srv := NewServer(dbConn)
    httpServer := &http.Server{
        // Addr:    net.JoinHostPort(config.Host, config.Port),
        Addr:    ":8080",
        Handler: srv,
    }

    if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		fmt.Fprintf(os.Stderr, "error listening and serving: %s\n", err)
        return err
	}
    return nil
}

func main() {
	fmt.Printf("Hello\n")
	var err error

	err = godotenv.Load(".env")
	if err != nil {
		log.Fatalf("main: Error loading .env file: %v", err)
	}

	// TODO: Test the token
    particleToken := os.Getenv("PARTICLE_TOKEN")
	if particleToken == "" {
		log.Fatalf("main: missing PARTICLE_TOKEN in .env file")
	}
    particle1, err := particle.NewParticle(particleToken)
    if err != nil {
        log.Fatalf("main: %+v", err)
    }

    dbConn, err := SetupDB("my.db3")
	if err != nil {
		log.Fatal("main: %w", err)
	}
	defer dbConn.Close()

    err = run(dbConn, particle1)
    if err != nil {
		log.Fatal("main: %w", err)
    }
}

