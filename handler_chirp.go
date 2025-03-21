package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/circuit-shell/http-server-go/internal/auth"
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
	// get the token from the request
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Error getting token", err)
		return
	}

	// validate the token
	userID, err := auth.ValidateJWT(token, cfg.serverSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Error validating token", err)
		return
	}

	decoder := json.NewDecoder(r.Body)
	params := chirpInput{}

	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error posting chirp, decoding params", err)
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
		UserID: userID,
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

func (cfg *apiConfig) handlerReadChirpById(w http.ResponseWriter, r *http.Request) {
	iDStr := r.PathValue("chirpID")
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
		respondWithError(w, http.StatusNotFound, "Error reading chirp", err)
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

func (cfg *apiConfig) handlerChirpsDelete(w http.ResponseWriter, r *http.Request) {
	chirpIDString := r.PathValue("chirpID")
	chirpID, err := uuid.Parse(chirpIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid chirp ID", err)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT", err)
		return
	}
	userID, err := auth.ValidateJWT(token, cfg.serverSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT", err)
		return
	}

	dbChirp, err := cfg.dbQueries.GetChirpsByID(r.Context(), chirpID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Couldn't get chirp", err)
		return
	}
	if dbChirp.UserID != userID {
		respondWithError(w, http.StatusForbidden, "You can't delete this chirp", err)
		return
	}

	err = cfg.dbQueries.DeleteChirp(r.Context(), chirpID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't delete chirp", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
