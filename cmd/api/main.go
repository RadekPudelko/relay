package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/RadekPudelko/relay/internal/database"
	"github.com/RadekPudelko/relay/internal/particle"
	"github.com/RadekPudelko/relay/internal/server"
	"github.com/RadekPudelko/relay/internal/config"

    "github.com/pelletier/go-toml"
)

func main() {
	fmt.Printf("Hello\n")
	err := run()
	if err != nil {
		log.Fatal("main: %w", err)
	}
}

func run() error {
    var err error
    defaultConfig := config.GetDefaultConfig()

    var myConfig *config.Config
	if _, err := os.Stat("config.toml"); err == nil {
        myConfig, err = config.LoadConfig("config.toml", &defaultConfig)
        log.Printf("run: loading config from config.toml")
        if err != nil {
            log.Fatalf("run: %+v", err)
        }
    } else {
        log.Printf("run: loading default config")
        myConfig = &defaultConfig
    }

    tomlData, err := toml.Marshal(myConfig)
    if err != nil {
        log.Fatalf("run: toml.Marshal: %v", err)
    }
    log.Printf("run: config=%s\n", string(tomlData))

	err = godotenv.Load(".env")
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

	dbConn, err := database.Setup(myConfig.Database.Filename, true)
	if err != nil {
		log.Fatal("run: %w", err)
	}
	defer dbConn.Close()

	go server.BackgroundTask(myConfig, dbConn, particle)
	err = server.Run(dbConn, myConfig.Server.Host, fmt.Sprintf("%d",myConfig.Server.Port))
	return nil
}
