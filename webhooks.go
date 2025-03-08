package main

import (
	"encoding/json"
	"net/http"

	"github.com/exy63/chirpy/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUpgradeUser(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	if apiKey, err := auth.GetAPIKey(r.Header); err != nil || apiKey != cfg.polkaKey {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	type Data struct {
		UserID string `json:"user_id"`
	}
	type ParsedRequest struct {
		Event string `json:"event"`
		Data  Data   `json:"data"`
	}
	var parsedRequest ParsedRequest

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&parsedRequest); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if parsedRequest.Event == "user.upgraded" {
		userUUID, err := uuid.Parse(parsedRequest.Data.UserID)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if err := cfg.dbQueries.UpgradeUser(r.Context(), userUUID); err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
	}

	w.WriteHeader(http.StatusNoContent)
}
