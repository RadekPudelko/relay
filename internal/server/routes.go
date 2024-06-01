package server

import (
	"database/sql"
	"net/http"
)

func addRoutes(
	mux *http.ServeMux,
	dbConn *sql.DB,
) {
	mux.Handle("GET /{$}", HandleGetRoot())
	mux.Handle("POST /api/relays", HandleCreateRelay(dbConn))
	mux.Handle("GET /api/relays/{id}", HandleGetRelay(dbConn))
	mux.Handle("DELETE /api/relays/{id}", HandleCancelRelay(dbConn))
}
