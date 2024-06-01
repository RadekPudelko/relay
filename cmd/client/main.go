package main

import (
	"fmt"
	"time"

	"github.com/joho/godotenv"

	"relay/client"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		fmt.Println(err)
	}

	client := client.NewClient(8080)
	err = client.Ping()
	if err != nil {
		fmt.Println(err)
	}

	deviceId := "device0"
	cloudFunction := "func0"
	argument := ""
	var desiredReturnCode *int = nil
	var scheduledTime *time.Time = nil

	id, err := client.CreateRelay(deviceId, cloudFunction, argument, desiredReturnCode, scheduledTime)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("Created relay %d\n", id)
	relay, err := client.GetRelay(id)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("%+v\n", relay)
}
