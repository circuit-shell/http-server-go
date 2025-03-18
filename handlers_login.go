package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/circuit-shell/http-server-go/internal/auth"
	"github.com/circuit-shell/http-server-go/internal/database"
)

var EXPIRES_IN_SECONDS = 3600

type AuthenticatedUser struct {
	User
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
}

type tokenResponse struct {
	Token string `json:"token"`
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
			ID:          user.ID,
			CreatedAt:   user.CreatedAt,
			UpdatedAt:   user.CreatedAt,
			Email:       user.Email,
			IsChirpyRed: user.IsChirpyRed,
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

func (cfg *apiConfig) handleRevoke(w http.ResponseWriter, r *http.Request) {
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

	// Revoke the refresh token
	err = cfg.dbQueries.RevokeRefreshTokens(r.Context(), refreshToken.UserID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error revoking refresh token", err)
		return
	}

	// respond with 204 No Content
	w.WriteHeader(http.StatusNoContent)

}
