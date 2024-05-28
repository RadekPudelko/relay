package server

import (
	"database/sql"
	"net/http"

	"pcfs/middleware"
)

func NewServer(dbConn *sql.DB) http.Handler {
	mux := http.NewServeMux()
	addRoutes(mux, dbConn)
	var handler http.Handler = mux
	handler = middleware.Logging(mux)
	return handler
}

