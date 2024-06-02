package server

import (
	"database/sql"
	"fmt"
	"net"
	"net/http"
	"os"

	"relay/internal/middleware"
)

func NewServer(db *sql.DB) http.Handler {
	mux := http.NewServeMux()
	addRoutes(mux, db)
	var handler http.Handler = mux
	handler = middleware.Logging(mux)
	return handler
}

func Run(db *sql.DB, host string, port string) error {
	srv := NewServer(db)
	httpServer := &http.Server{
		Addr:    net.JoinHostPort(host, port),
		Handler: srv,
	}

	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		fmt.Fprintf(os.Stderr, "error listening and serving: %s\n", err)
		return err
	}
	return nil
}
