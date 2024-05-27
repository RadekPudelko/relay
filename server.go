package main
import(
    "net/http"
	"database/sql"

	"pcfs/middleware"
)

func NewServer(dbConn *sql.DB) http.Handler {
	mux := http.NewServeMux()
	addRoutes(mux, dbConn)
	var handler http.Handler = mux
    handler = middleware.Logging(mux)
	return handler
}

