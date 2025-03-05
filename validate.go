package main

import (
	"encoding/json"
	"net/http"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

type SuccessResponse struct {
	Valid bool `json:"valid"`
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
		res, _ := json.Marshal(ErrorResponse{Error: "Something went wrong"})
		w.WriteHeader(http.StatusBadRequest)
		w.Write(res)
		return
	}

	if len(chirp.Body) > maxChirpLength {
		res, _ := json.Marshal(ErrorResponse{Error: "Chirp is too long"})
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(res))
		return
	}

	res, _ := json.Marshal(SuccessResponse{Valid: true})
	w.WriteHeader(http.StatusOK)
	w.Write(res)
}
