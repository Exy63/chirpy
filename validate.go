package main

import (
	"encoding/json"
	"net/http"
	"strings"
)

type SuccessResponse struct {
	CleanedBody string `json:"cleaned_body"`
}

const maxChirpLength = 140

func handlerValidate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	defer r.Body.Close()

	type Chirp struct {
		Body string `json:"body"`
	}
	var chirp Chirp

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&chirp); err != nil {
		respondWithError(w, http.StatusBadRequest, "Something went wrong")
		return
	}

	if len(chirp.Body) > maxChirpLength {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long")
		return
	}

	cleaned := getCleanedMsg(chirp.Body)

	respondWithJSON(w, http.StatusOK, SuccessResponse{CleanedBody: cleaned})
}

func getCleanedMsg(msg string) string {
	badWords := map[string]struct{}{
		"kerfuffle": {},
		"sharbert":  {},
		"fornax":    {},
	}

	words := strings.Split(msg, " ")
	for i, word := range words {
		loweredWord := strings.ToLower(word)
		if _, ok := badWords[loweredWord]; ok {
			words[i] = "****"
		}
	}

	return strings.Join(words, " ")
}
