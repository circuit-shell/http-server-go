package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

type userInput struct {
	Email string `json:"email"`
}

func (cfg *apiConfig) handlerCreateUser(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	userParams := userInput{}
	err := decoder.Decode(&userParams)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error decoding user params", err)
		return
	}
	user, err := cfg.dbQueries.CreateUser(r.Context(), userParams.Email)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Error creating db", err)
		return
	}
	respondWithJSON(w, http.StatusCreated, User{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.CreatedAt,
		Email:     user.Email,
	})
}

func (cfg *apiConfig) handlerReadUsers(w http.ResponseWriter, r *http.Request) {
	users, err := cfg.dbQueries.GetUsers(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error reading users", err)
		return
	}
	respondWithJSON(w, http.StatusOK, users)
}
