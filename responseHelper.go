package main

import (
	"encoding/json"
	"net/http"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	res, _ := json.Marshal(ErrorResponse{Error: msg})
	w.WriteHeader(code)
	w.Write(res)
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	res, err := json.Marshal(payload)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Something went wrong")
	}
	w.WriteHeader(code)
	w.Write(res)
}
