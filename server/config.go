package server

import "time"

type Config struct {
	Host              string
	Port              string
	MaxRoutines       int
	PingRetryDuration time.Duration
	CFRetryDuration   time.Duration
	TaskLimit         int
	MaxRetries        int
}
