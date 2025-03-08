package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/exy63/chirpy/internal/auth"
	"github.com/exy63/chirpy/internal/database"

	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerCreateUser(w http.ResponseWriter, r *http.Request) {
	type Response struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Email     string    `json:"email"`
	}

	w.Header().Set("Content-Type", "application/json")
	defer r.Body.Close()

	type ParsedRequest struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	var parsedRequest ParsedRequest

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&parsedRequest); err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	if parsedRequest.Email == "" || parsedRequest.Password == "" {
		respondWithError(w, http.StatusBadRequest, "Email and Password are required")
		return
	}

	hashedPassword, err := auth.HashPassword(parsedRequest.Password)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	hashedPasswordNull := sql.NullString{
		String: hashedPassword,
		Valid:  true,
	}

	params := database.CreateUserParams{
		Email:          parsedRequest.Email,
		HashedPassword: hashedPasswordNull,
	}

	createdUser, err := cfg.dbQueries.CreateUser(r.Context(), params)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	userResponse := Response{
		ID:        createdUser.ID,
		CreatedAt: createdUser.CreatedAt,
		UpdatedAt: createdUser.UpdatedAt,
		Email:     createdUser.Email,
	}

	respondWithJSON(w, http.StatusCreated, userResponse)
}

func (cfg *apiConfig) handlerLoginUser(w http.ResponseWriter, r *http.Request) {
	type Response struct {
		ID           uuid.UUID `json:"id"`
		CreatedAt    time.Time `json:"created_at"`
		UpdatedAt    time.Time `json:"updated_at"`
		Email        string    `json:"email"`
		Token        string    `json:"token"`
		RefreshToken string    `json:"refresh_token"`
	}

	w.Header().Set("Content-Type", "application/json")
	defer r.Body.Close()

	type ParsedRequest struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	var parsedRequest ParsedRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&parsedRequest); err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	if parsedRequest.Email == "" || parsedRequest.Password == "" {
		respondWithError(w, http.StatusBadRequest, "Email and Password are required")
		return
	}

	userFromDb, err := cfg.dbQueries.GetUserByEmail(r.Context(), parsedRequest.Email)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password")
		return
	}

	if err := auth.CheckPasswordHash(parsedRequest.Password, userFromDb.HashedPassword.String); err != nil {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password")
		return
	}

	accessToken, err := auth.MakeJWT(userFromDb.ID, cfg.jwtSecret, time.Duration(1*time.Hour))
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create an access token")
		return
	}

	refreshToken := auth.MakeRefreshToken()
	const sixtyDaysInHours = 1440
	params := database.CreateRefreshTokenParams{
		Token:     refreshToken,
		UserID:    userFromDb.ID,
		ExpiresAt: time.Now().Add(time.Duration(sixtyDaysInHours) * time.Hour),
	}
	if _, err := cfg.dbQueries.CreateRefreshToken(r.Context(), params); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create a refresh token")
		return
	}

	userResponse := Response{
		ID:           userFromDb.ID,
		CreatedAt:    userFromDb.CreatedAt,
		UpdatedAt:    userFromDb.UpdatedAt,
		Email:        userFromDb.Email,
		Token:        accessToken,
		RefreshToken: refreshToken,
	}

	respondWithJSON(w, http.StatusOK, userResponse)
}
