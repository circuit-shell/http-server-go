package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/circuit-shell/http-server-go/internal/auth"
	"github.com/circuit-shell/http-server-go/internal/database"
	"github.com/google/uuid"
)

var EXPIRES_IN_SECONDS = 3600

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

type AuthenticatedUser struct {
	User
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
}

type tokenResponse struct {
	Token string `json:"token"`
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
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.CreatedAt,
		Email:     user.Email,
	})
}

func (cfg *apiConfig) handleLogin(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	userParams := userInput{}

	// Decode the user input
	err := decoder.Decode(&userParams)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error decoding user params", err)
		return
	}

	// Get the user from the database
	user, err := cfg.dbQueries.GetUserByEmail(r.Context(), userParams.Email)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid user", err)
		return
	}

	// Check the password
	if !auth.CheckPasswordHash(userParams.Password, user.HashedPassword) {
		respondWithError(w, http.StatusUnauthorized, "Invalid password", err)
		return
	}

	// Generate the access token
	token, err := auth.MakeJWT(user.ID, cfg.serverSecret, time.Duration(EXPIRES_IN_SECONDS)*time.Second)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error generating token", err)
		return
	}

	// generate the resfresh token
	refreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error generating refresh token", err)
		return
	}

	// Save the refresh token in the database
	_, err = cfg.dbQueries.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
		UserID: user.ID,
		Token:  refreshToken,
	})

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error saving refresh token", err)
		return
	}

	respondWithJSON(w, http.StatusOK, AuthenticatedUser{
		User: User{
			ID:        user.ID,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.CreatedAt,
			Email:     user.Email,
		},
		Token:        token,
		RefreshToken: refreshToken,
	})

}

func (cfg *apiConfig) handleRefresh(w http.ResponseWriter, r *http.Request) {
	// get the token from the request
	inputToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Error getting refresh token", err)
		return
	}

	// look up the refresh token in the database

	refreshToken, err := cfg.dbQueries.GetUserFromRefreshToken(r.Context(), inputToken)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid refresh token", err)
		return
	}

	// Generate the refreshed access token
	token, err := auth.MakeJWT(refreshToken.UserID, cfg.serverSecret, time.Duration(EXPIRES_IN_SECONDS)*time.Second)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error generating token", err)
		return
	}

	respondWithJSON(w, http.StatusOK, tokenResponse{
		Token: token,
	})
}

func (cfg *apiConfig) handlerReadUsers(w http.ResponseWriter, r *http.Request) {
	users, err := cfg.dbQueries.GetUsers(r.Context())
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Error reading users", err)
		return
	}
	respondWithJSON(w, http.StatusOK, users)
}

func (cfg *apiConfig) handlerReadUser(w http.ResponseWriter, r *http.Request) {

	userIDStr := r.PathValue("id")
	if userIDStr == "" {
		respondWithError(w, http.StatusBadRequest, "Missing user ID in path", nil)
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error need a user ID", err)
		return
	}
	user, err := cfg.dbQueries.GetUserByID(r.Context(), userID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error reading user", err)
		return
	}

	respondWithJSON(w, http.StatusOK, User{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.CreatedAt,
		Email:     user.Email,
	})
}
