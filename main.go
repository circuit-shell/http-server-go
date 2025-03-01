package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/circuit-shell/http-server-go/internal/database"
	"github.com/joho/godotenv"

	_ "github.com/lib/pq"
)

func main() {
	apiCfg := &apiConfig{}

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading envs", err)
	}
	dbURL := os.Getenv("DB_URL")

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}

	dbQueries := database.New(db)
	apiCfg.dbQueries = dbQueries
  apiCfg.platform = os.Getenv("PLATFORM")

	const filepathRoot = "."
	const port = "8080"
	mux := http.NewServeMux()

	mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot)))))

	mux.HandleFunc("GET /api/healthz", handlerReadiness)
	mux.HandleFunc("POST /api/validate_chirp", handlerChirpLen)
	mux.HandleFunc("POST /api/users", apiCfg.handlerUsers)

	mux.HandleFunc("GET /admin/metrics", apiCfg.handlerMetrics)
	mux.HandleFunc("POST /admin/reset", apiCfg.handlerMetricsReset)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}
	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Printf("Serving http://localhost:%v", port)
	log.Fatal(srv.ListenAndServe())
}
