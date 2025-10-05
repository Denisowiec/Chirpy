package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

type ChirpMessage struct {
	Body string `json:"body"`
}

type ChirpValidResponse struct {
	CleanedBody string `json:"cleaned_body"`
}

func replaceProfane(s string) string {
	grawlix := "****"
	profanities := []string{
		"kerfuffle",
		"sharbert",
		"fornax",
	}
	split := strings.Fields(s)
	for _, prof := range profanities {
		for i, word := range split {
			if strings.ToLower(word) == prof {
				split[i] = grawlix
			}
		}
	}

	return strings.Join(split, " ")
}

func handlerValidateChirp(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	chirp := ChirpMessage{}

	// Regardless what happens, the response will be json
	w.Header().Set("Content-Type", "application/json")

	// decoding the incoming chirp into an appropriate struct
	if err := decoder.Decode(&chirp); err != nil {
		log.Printf("error decoding parameters: %s", err)

		// If there's an error, we still send out a response
		w.WriteHeader(http.StatusInternalServerError) // Code 500

		errText := generateErrorResp("Something went wrong")
		w.Write(errText)
		return
	}

	// Testing if the chirp is too long
	if len(chirp.Body) == 0 {
		w.WriteHeader(http.StatusBadRequest) // Code 400

		errBody := ChirpErrorResponse{
			Error: "Chirp malformed",
		}
		dat, err := json.Marshal(errBody)
		if err != nil {
			log.Printf("Error marshalling JSON: %s", err)
			return
		}
		w.Write(dat)
		return

	} else if len(chirp.Body) > 140 {
		w.WriteHeader(http.StatusBadRequest) // Code 400

		errBody := ChirpErrorResponse{
			Error: "Chirp is too long",
		}
		dat, err := json.Marshal(errBody)
		if err != nil {
			log.Printf("Error marshalling JSON: %s", err)
			return
		}
		w.Write(dat)
		return
	} else {
		w.WriteHeader(http.StatusOK) // Code 200

		respBody := ChirpValidResponse{
			CleanedBody: replaceProfane(chirp.Body),
		}

		dat, err := json.Marshal(respBody)
		if err != nil {
			log.Printf("Error marshalling JSON %s", err)
			return
		}
		w.Write(dat)
	}

}
