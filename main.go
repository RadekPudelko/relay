package main

import (
	"database/sql"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"relay/db"
	"relay/particle"
	"relay/server"
)

func main() {
	fmt.Printf("Hello\n")
	err := run()
	if err != nil {
		log.Fatal("main: %w", err)
	}
}

func run() error {
	config := server.Config{
		Host:        "localhost",
		Port:        "8080",
		MaxRoutines: 2,
		TaskLimit:   10,
		MaxRetries:  3,
	}

	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("run: Error loading .env file: %v", err)
	}

	particleToken := os.Getenv("PARTICLE_TOKEN")
	if particleToken == "" {
		log.Fatalf("run: missing PARTICLE_TOKEN in .env file")
	}
	particle, err := particle.NewParticle(particleToken)
	if err != nil {
		log.Fatalf("run: %+v", err)
	}

	dbPath := os.Getenv("DB")
	if particleToken == "" {
		log.Fatalf("run: missing PARTICLE_TOKEN in .env file")
	}
	dbConn, err := SetupDB(dbPath)
	if err != nil {
		log.Fatal("run: %w", err)
	}
	defer dbConn.Close()

	err = Run(config, dbConn, particle)
	return nil
}

// TODO: move this somewhere else
func SetupDB(path string) (*sql.DB, error) {
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

func Run(
	config server.Config,
	dbConn *sql.DB,
	particle particle.ParticleAPI,
) error {

	go server.BackgroundTask(config, dbConn, particle)

	srv := server.NewServer(dbConn)
	httpServer := &http.Server{
		Addr:    net.JoinHostPort(config.Host, config.Port),
		Handler: srv,
	}

	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		fmt.Fprintf(os.Stderr, "error listening and serving: %s\n", err)
		return err
	}
	return nil
}
