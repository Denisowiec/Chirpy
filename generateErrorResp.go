package main

import (
	"encoding/json"
	"log"
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
