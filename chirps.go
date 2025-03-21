package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/exy63/chirpy/internal/auth"
	"github.com/exy63/chirpy/internal/database"
	"github.com/google/uuid"
)

type ChirpResponse struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    string    `json:"user_id"`
}

func (cfg *apiConfig) handlerGetChirp(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	defer r.Body.Close()

	id := r.PathValue("id")
	uuid, err := uuid.Parse(id)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	chirp, err := cfg.dbQueries.GetChirp(r.Context(), uuid)
	if err != nil {
		respondWithError(w, http.StatusNotFound, err.Error())
		return
	}

	chirpResponse := ChirpResponse{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID.String(),
	}

	respondWithJSON(w, http.StatusOK, chirpResponse)
}

func (cfg *apiConfig) handlerGetChirps(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	defer r.Body.Close()

	authorID := r.URL.Query().Get("author_id")
	userIDParam := uuid.NullUUID{}
	if authorID != "" {
		parsedUUID, err := uuid.Parse(authorID)
		if err == nil {
			userIDParam = uuid.NullUUID{UUID: parsedUUID, Valid: true}
		}
	}

	chirps, err := cfg.dbQueries.GetChirps(r.Context(), userIDParam)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	chirpResponse := make([]ChirpResponse, len(chirps))
	for i, chirp := range chirps {
		chirpResponse[i] = ChirpResponse{
			ID:        chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body:      chirp.Body,
			UserID:    chirp.UserID.String(),
		}
	}

	if r.URL.Query().Get("sort") == "desc" {
		sort.Slice(chirpResponse, func(i, j int) bool {
			return chirpResponse[i].CreatedAt.After(chirpResponse[j].CreatedAt)
		})
	}

	respondWithJSON(w, http.StatusOK, chirpResponse)
}

func (cfg *apiConfig) handlerCreateChirp(w http.ResponseWriter, r *http.Request) {
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
		Body   string    `json:"body"`
		UserID uuid.UUID `json:"user_id"`
	}
	var req Request

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body format")
		return
	}

	cleanedBody, err := getCleanedBody(req.Body)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	params := database.CreateChirpParams{
		Body:   cleanedBody,
		UserID: UserID,
	}
	chirp, err := cfg.dbQueries.CreateChirp(r.Context(), params)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	chirpResponse := ChirpResponse{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID.String(),
	}

	respondWithJSON(w, http.StatusCreated, chirpResponse)
}

func getCleanedBody(msg string) (string, error) {
	const maxChirpLength = 140
	badWords := map[string]struct{}{
		"kerfuffle": {},
		"sharbert":  {},
		"fornax":    {},
	}

	if len(msg) > maxChirpLength {
		return "", errors.New("chirp is too long")
	}

	words := strings.Split(msg, " ")
	for i, word := range words {
		loweredWord := strings.ToLower(word)
		if _, ok := badWords[loweredWord]; ok {
			words[i] = "****"
		}
	}

	return strings.Join(words, " "), nil
}

func (cfg *apiConfig) handlerDeleteChirp(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	providedToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	UserID, err := auth.ValidateJWT(providedToken, cfg.jwtSecret)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	id := r.PathValue("id")
	uuid, err := uuid.Parse(id)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	chirp, err := cfg.dbQueries.GetChirp(r.Context(), uuid)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if chirp.UserID != UserID {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	if err := cfg.dbQueries.DeleteChirp(r.Context(), chirp.ID); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
