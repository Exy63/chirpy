package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type UserResponse struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

func (cfg *apiConfig) handlerCreateUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	defer r.Body.Close()

	type ParsedRequest struct {
		Email string `json:"email"`
	}
	var parsedRequest ParsedRequest

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&parsedRequest); err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	createdUser, err := cfg.dbQueries.CreateUser(r.Context(), parsedRequest.Email)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	userResponse := UserResponse{
		ID:        createdUser.ID,
		CreatedAt: createdUser.CreatedAt,
		UpdatedAt: createdUser.UpdatedAt,
		Email:     createdUser.Email,
	}

	respondWithJSON(w, http.StatusCreated, userResponse)
}
