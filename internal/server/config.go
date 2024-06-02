package server

import "time"

// Add config loading from file
type Config struct {
	Host              string
	Port              string
	MaxRoutines       int
	PingRetryDuration time.Duration
	CFRetryDuration   time.Duration
	RelayLimit        int
	MaxRetries        int
}

// func NewConfig(host, port string, maxRoutines int, pingRetryDuration, cfRetryDuration time.Duration, relayLimit, MaxRetries int) (Config) {
//     return Config{
//         Host: host,
//         Port: port,
//         MaxRoutines: ,
//     }
// }
