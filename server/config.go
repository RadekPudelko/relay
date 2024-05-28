package server

type Config struct {
	Host        string
	Port        string
	MaxRoutines int
	TaskLimit   int
	MaxRetries  int
}

