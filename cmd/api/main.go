package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"relay/internal/database"
	"relay/internal/particle"
	"relay/internal/server"
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
		Host:              "localhost",
		Port:              "8080",
		MaxRoutines:       3,
		RelayLimit:        10,
		PingRetryDuration: 600 * time.Second,
		CFRetryDuration:   600 * time.Second,
		MaxRetries:        3,
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
	dbConn, err := database.Setup(dbPath, true)
	if err != nil {
		log.Fatal("run: %w", err)
	}
	defer dbConn.Close()

	go server.BackgroundTask(config, dbConn, particle)
	err = server.Run(config, dbConn)
	return nil
}

