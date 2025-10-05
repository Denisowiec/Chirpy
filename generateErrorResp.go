package main

import (
	"encoding/json"
	"log"
	"net/http"
)

type ChirpErrorResponse struct {
	Error string `json:"error"`
}

func generateErrorResp(s string) []byte {
	errBody := ChirpErrorResponse{
		Error: s,
	}
	dat, err := json.Marshal(errBody)
	if err != nil {
		log.Fatalf("Error marshalling JSON: %s", err)
	}
	return dat
}

func respondError(w http.ResponseWriter, message string, errorCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(errorCode)
	dat := generateErrorResp(message)
	w.Write(dat)
}
