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

type UserResponse struct {
	ID          uuid.UUID `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Email       string    `json:"email"`
	IsChirpyRed bool      `json:"is_chirpy_red"`
}

func (cfg *apiConfig) handlerCreateUser(w http.ResponseWriter, r *http.Request) {
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

	userResponse := UserResponse{
		ID:          createdUser.ID,
		CreatedAt:   createdUser.CreatedAt,
		UpdatedAt:   createdUser.UpdatedAt,
		Email:       createdUser.Email,
		IsChirpyRed: createdUser.IsChirpyRed.Bool,
	}

	respondWithJSON(w, http.StatusCreated, userResponse)
}

func (cfg *apiConfig) handlerUpdateUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	defer r.Body.Close()

	providedToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "You must provide a token")
		return
	}
	UserID, err := auth.ValidateJWT(providedToken, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	type Request struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	var req Request

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	userFromDb, err := cfg.dbQueries.GetUser(r.Context(), UserID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, err.Error())
		return
	}

	emailForUpdate := userFromDb.Email
	if req.Email != "" {
		emailForUpdate = req.Email
	}
	passwordForUpdate := userFromDb.HashedPassword
	if req.Password != "" {
		hashedPassword, err := auth.HashPassword(req.Password)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}

		passwordForUpdate = sql.NullString{
			String: hashedPassword,
			Valid:  true,
		}
	}

	params := database.UpdateUserParams{
		ID:             UserID,
		Email:          emailForUpdate,
		HashedPassword: passwordForUpdate,
	}

	updatedUser, err := cfg.dbQueries.UpdateUser(r.Context(), params)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	userResponse := UserResponse{
		ID:          updatedUser.ID,
		CreatedAt:   updatedUser.CreatedAt,
		UpdatedAt:   updatedUser.UpdatedAt,
		Email:       updatedUser.Email,
		IsChirpyRed: updatedUser.IsChirpyRed.Bool,
	}

	respondWithJSON(w, http.StatusOK, userResponse)
}

func (cfg *apiConfig) handlerLoginUser(w http.ResponseWriter, r *http.Request) {
	type LoginUserResponse struct {
		ID           uuid.UUID `json:"id"`
		CreatedAt    time.Time `json:"created_at"`
		UpdatedAt    time.Time `json:"updated_at"`
		Email        string    `json:"email"`
		Token        string    `json:"token"`
		RefreshToken string    `json:"refresh_token"`
		IsChirpyRed  bool      `json:"is_chirpy_red"`
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

	userResponse := LoginUserResponse{
		ID:           userFromDb.ID,
		CreatedAt:    userFromDb.CreatedAt,
		UpdatedAt:    userFromDb.UpdatedAt,
		Email:        userFromDb.Email,
		Token:        accessToken,
		RefreshToken: refreshToken,
		IsChirpyRed:  userFromDb.IsChirpyRed.Bool,
	}

	respondWithJSON(w, http.StatusOK, userResponse)
}
