package server

import (
	"database/sql"
	"fmt"
	"net"
	"net/http"
	"os"

	"relay/middleware"
	"relay/particle"
)

func NewServer(dbConn *sql.DB) http.Handler {
	mux := http.NewServeMux()
	addRoutes(mux, dbConn)
	var handler http.Handler = mux
	handler = middleware.Logging(mux)
	return handler
}

func Run(
	config Config,
	dbConn *sql.DB,
	particle particle.ParticleAPI,
) error {
	go BackgroundTask(config, dbConn, particle)

	srv := NewServer(dbConn)
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
