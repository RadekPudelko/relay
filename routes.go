package main

import (
	"database/sql"
	"net/http"
)

func addRoutes(
	mux *http.ServeMux,
	dbConn *sql.DB,
) {
	mux.Handle("GET /{$}", HandleGetRoot())
	mux.Handle("POST /api/tasks", HandleCreateTask(dbConn))
	mux.Handle("GET /api/tasks/{id}", HandleGetTask(dbConn))
}
