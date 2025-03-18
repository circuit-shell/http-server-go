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

	apiCfg.dbQueries = database.New(db)
	apiCfg.platform = os.Getenv("PLATFORM")
	apiCfg.serverSecret = os.Getenv("SERVER_SECRET")

	log.Printf("Connected to database: %s", dbURL)
	log.Printf("Server secret: %s", os.Getenv("SERVER_SECRET"))
	log.Printf("Platform: %s", os.Getenv("PLATFORM"))

	const filepathRoot = "."
	const port = "8080"
	mux := http.NewServeMux()

	mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot)))))

	mux.HandleFunc("GET /api/healthz", handlerReadiness)

	mux.HandleFunc("GET /admin/metrics", apiCfg.handlerMetrics)
	mux.HandleFunc("POST /admin/reset", apiCfg.handlerMetricsReset)

	mux.HandleFunc("POST /api/login", apiCfg.handleLogin)
	mux.HandleFunc("POST /api/refresh", apiCfg.handleRefresh)
	mux.HandleFunc("POST /api/revoke", apiCfg.handleRevoke)

	mux.HandleFunc("POST /api/chirps", apiCfg.handlerCreateChirp)
	mux.HandleFunc("DELETE /api/chirps/{chirpID}", apiCfg.handlerChirpsDelete)
	mux.HandleFunc("GET /api/chirps", apiCfg.handlerReadChirps)
	mux.HandleFunc("GET /api/chirps/{chirpID}", apiCfg.handlerReadChirpById)

	mux.HandleFunc("POST /api/users", apiCfg.handlerCreateUser)
  mux.HandleFunc("PUT /api/users", apiCfg.handlerUpdateUser)
	mux.HandleFunc("GET /api/users", apiCfg.handlerReadUsers)
	mux.HandleFunc("GET /api/users/{id}", apiCfg.handlerReadUser)


	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}
	log.Printf("Serving http://localhost:%v", port)
	log.Fatal(srv.ListenAndServe())
}
