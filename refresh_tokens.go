package main

import (
	"net/http"
	"time"

	"github.com/exy63/chirpy/internal/auth"
)

func (cfg *apiConfig) handlerRefreshToken(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	defer r.Body.Close()

	providedToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "You must provide a token")
		return
	}

	refreshToken, err := cfg.dbQueries.GetRefreshToken(r.Context(), providedToken)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	accessToken, err := auth.MakeJWT(refreshToken.UserID, cfg.jwtSecret, time.Duration(1)*time.Hour)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create an access token")
		return
	}

	type Response struct {
		Token string `json:"token"`
	}

	res := Response{
		Token: accessToken,
	}

	respondWithJSON(w, http.StatusOK, res)
}

func (cfg *apiConfig) handlerRevokeToken(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	providedToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "You must provide a token")
		return
	}

	refreshToken, err := cfg.dbQueries.GetRefreshToken(r.Context(), providedToken)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if err := cfg.dbQueries.RevokeRefreshToken(r.Context(), refreshToken.Token); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
