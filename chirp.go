package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/circuit-shell/http-server-go/internal/database"
	"github.com/google/uuid"
)

const MAX_CHIRP_LENGTH = 140

type chirpInput struct {
	Body   string    `json:"body"`
	UserID uuid.UUID `json:"user_id"`
}

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

func (cfg *apiConfig) handlerCreateChirp(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	params := chirpInput{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error decoding params", err)
		return
	}
	if len(params.Body) > MAX_CHIRP_LENGTH {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long", nil)
		return
	}

	cleaned_body := censorProfanity(params.Body)

	// Add error handling for database operation
	chirp, err := cfg.dbQueries.CreateChirp(r.Context(), database.CreateChirpParams{
		Body:   cleaned_body,
		UserID: params.UserID, // Changed UserId to UserID
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error creating chirp", err)
		return
	}

	respondWithJSON(w, http.StatusCreated, Chirp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.CreatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	})

}

func (cfg *apiConfig) handlerReadChirps(w http.ResponseWriter, r *http.Request) {
	chirps, err := cfg.dbQueries.GetChirps(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error reading chirps", err)
		return
	}

	formatted_chirps := []Chirp{}

	for _, chirp := range chirps {
		formatted_chirps = append(formatted_chirps, Chirp{
			ID:        chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body:      chirp.Body,
			UserID:    chirp.UserID,
		})

	}
	respondWithJSON(w, http.StatusOK, formatted_chirps)
}

func (cfg *apiConfig) handlerReadChirp(w http.ResponseWriter, r *http.Request) {
	iDStr := r.PathValue("id")
	if iDStr == "" {
		respondWithError(w, http.StatusBadRequest, "Missing chirp ID in path", nil)
		return
	}

	userID, err := uuid.Parse(iDStr)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error need a chirp ID", err)
		return
	}
	chirp, err := cfg.dbQueries.GetChirpsByID(r.Context(), userID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error reading chirp", err)
		return
	}

	respondWithJSON(w, http.StatusOK, Chirp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	})
}
