package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/circuit-shell/http-server-go/internal/auth"
	"github.com/circuit-shell/http-server-go/internal/database"
	"github.com/google/uuid"
)

type User struct {
	ID          uuid.UUID `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Email       string    `json:"email"`
	IsChirpyRed bool      `json:"is_chirpy_red"`
}

type userInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (cfg *apiConfig) handlerCreateUser(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	userParams := userInput{}
	err := decoder.Decode(&userParams)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error decoding user params", err)
		return
	}

	hashedPw, err := auth.HashPassword(userParams.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error hashing password", err)
		return
	}

	user, err := cfg.dbQueries.CreateUser(r.Context(), database.CreateUserParams{
		Email:          userParams.Email,
		HashedPassword: hashedPw,
	})
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Error creating user", err)
		return
	}
	respondWithJSON(w, http.StatusCreated, User{
		ID:          user.ID,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.CreatedAt,
		Email:       user.Email,
		IsChirpyRed: user.IsChirpyRed,
	})
}

func (cfg *apiConfig) handlerUpdateUser(w http.ResponseWriter, r *http.Request) {

	// get the token from the request
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Error getting token", err)
		return
	}

	// validate the token
	userID, err := auth.ValidateJWT(token, cfg.serverSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Error validating token", err)
		return
	}

	decoder := json.NewDecoder(r.Body)
	userParams := userInput{}

	err = decoder.Decode(&userParams)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error decoding user params", err)
		return
	}

	hashedPw, err := auth.HashPassword(userParams.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error hashing password", err)
		return

	}

	user, err := cfg.dbQueries.UpdateUser(r.Context(), database.UpdateUserParams{
		ID:             userID,
		Email:          userParams.Email,
		HashedPassword: hashedPw,
	})

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error updating user", err)
		return
	}

	respondWithJSON(w, http.StatusOK, User{
		ID:          user.ID,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.CreatedAt,
		IsChirpyRed: user.IsChirpyRed,
		Email:       user.Email,
	})

}

// func (cfg *apiConfig) handlerReadUsers(w http.ResponseWriter, r *http.Request) {
// 	users, err := cfg.dbQueries.GetUsers(r.Context())
// 	if err != nil {
// 		respondWithError(w, http.StatusNotFound, "Error reading users", err)
// 		return
// 	}
// 	respondWithJSON(w, http.StatusOK, users)
// }

// func (cfg *apiConfig) handlerReadUser(w http.ResponseWriter, r *http.Request) {

// 	userIDStr := r.PathValue("id")
// 	if userIDStr == "" {
// 		respondWithError(w, http.StatusBadRequest, "Missing user ID in path", nil)
// 		return
// 	}

// 	userID, err := uuid.Parse(userIDStr)
// 	if err != nil {
// 		respondWithError(w, http.StatusInternalServerError, "Error need a user ID", err)
// 		return
// 	}
// 	user, err := cfg.dbQueries.GetUserByID(r.Context(), userID)
// 	if err != nil {
// 		respondWithError(w, http.StatusInternalServerError, "Error reading user", err)
// 		return
// 	}

// 	respondWithJSON(w, http.StatusOK, User{
// 		ID:        user.ID,
// 		CreatedAt: user.CreatedAt,
// 		UpdatedAt: user.CreatedAt,
// 		Email:     user.Email,
// 	})
// }
