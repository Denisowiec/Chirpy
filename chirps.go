package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/Denisowiec/Chirpy/internal/auth"
	"github.com/Denisowiec/Chirpy/internal/database"
	"github.com/google/uuid"
)

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

func (cfg *apiConfig) handlerPostChirp(w http.ResponseWriter, r *http.Request) {
	type chirpMinimal struct {
		Body   string    `json:"body"`
		UserID uuid.UUID `json:"user_id"`
	}
	decoder := json.NewDecoder(r.Body)
	chirpInput := chirpMinimal{}

	// Regardless what happens, the response will be json
	w.Header().Set("Content-Type", "application/json")

	// decoding the incoming chirp into an appropriate struct
	if err := decoder.Decode(&chirpInput); err != nil {
		respondError(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	// User authentification
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("Error extracting jwt from header: %s", err)
		respondError(w, "Authentification failed", http.StatusUnauthorized)
		return
	}
	inUID, err := auth.ValidateJWT(token, cfg.jwtSecretCode)
	if err != nil {
		respondError(w, "Authentification failed", http.StatusUnauthorized)
		return
	}

	// Testing if the chirp is too long
	if len(chirpInput.Body) == 0 {
		respondError(w, "Chirp malformed", http.StatusBadRequest)
		return

	} else if len(chirpInput.Body) > 140 {
		respondError(w, "Chirp is too long", http.StatusBadRequest)
		return
	} else {

		chirpInput.Body = replaceProfane(chirpInput.Body)

		ccparams := database.CreateChirpParams{
			Body:   chirpInput.Body,
			UserID: inUID,
		}

		chirp, err := cfg.db.CreateChirp(r.Context(), ccparams)
		if err != nil {
			log.Printf("Error putting chirp into database: %v", err)
			respondError(w, "Couldn't process the chirp into database", http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusCreated) // Code 201

		dat, err := json.Marshal(chirp)
		if err != nil {
			log.Printf("Error marshalling JSON %s", err)
			return
		}
		w.Write(dat)
	}

}

func (cfg *apiConfig) handlerGetChirps(w http.ResponseWriter, r *http.Request) {
	chirps, err := cfg.db.GetChirps(r.Context())
	if err != nil {
		log.Printf("Error getting chirps from the database: %s", err)
		respondError(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	dat, err := json.Marshal(chirps)
	if err != nil {
		log.Printf("Error marshalling data")
		respondError(w, "Internal server error", http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK) // Code 200
	w.Write(dat)
}

func (cfg *apiConfig) handlerGetChirp(w http.ResponseWriter, r *http.Request) {
	reqId, err := uuid.Parse(r.PathValue("chirpid"))
	if err != nil {
		log.Printf("Error parsing chirp id: %s", err)
		respondError(w, "Error processing request", http.StatusInternalServerError)
		return
	}

	chirp, err := cfg.db.GetChirpById(r.Context(), reqId)
	if err != nil {
		log.Printf("Error getting chirp from database: %s", err)
		respondError(w, "Chirp not found", http.StatusNotFound)
		return
	}

	dat, err := json.Marshal(chirp)
	if err != nil {
		log.Printf("Error marshalling data: %s", err)
		respondError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(dat)
}

func (cfg *apiConfig) handlerDeleteChirp(w http.ResponseWriter, r *http.Request) {
	// User authentification
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("Error extracting jwt from header: %s", err)
		respondError(w, "Authentification failed", http.StatusUnauthorized)
		return
	}
	inUID, err := auth.ValidateJWT(token, cfg.jwtSecretCode)
	if err != nil {
		respondError(w, "Authentification failed", http.StatusUnauthorized)
		return
	}

	reqId, err := uuid.Parse(r.PathValue("chirpid"))
	if err != nil {
		log.Printf("Error parsing chirp id: %s", err)
		respondError(w, "Error processing request", http.StatusInternalServerError)
		return
	}

	chirp, err := cfg.db.GetChirpById(r.Context(), reqId)
	if err != nil {
		respondError(w, "Chirp not found", http.StatusNotFound)
		return
	}
	if chirp.UserID != inUID {
		respondError(w, "Operation unauthorized", http.StatusForbidden)
		return
	}

	delChirpParams := database.DeleteChirpParams{
		ID:     reqId,
		UserID: inUID,
	}
	_, err = cfg.db.DeleteChirp(r.Context(), delChirpParams)
	if err != nil {
		log.Printf("Error deleting chirp: %s", err)
		respondError(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)
}
