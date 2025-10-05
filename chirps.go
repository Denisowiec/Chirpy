package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/Denisowiec/Chirpy/internal/database"
	"github.com/google/uuid"
)

type chirpMinimal struct {
	Body   string    `json:"body"`
	UserID uuid.UUID `json:"user_id"`
}

type chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
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

func (cfg *apiConfig) handlerPostChirp(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	chirp := chirpMinimal{}

	// Regardless what happens, the response will be json
	w.Header().Set("Content-Type", "application/json")

	// decoding the incoming chirp into an appropriate struct
	if err := decoder.Decode(&chirp); err != nil {
		respondError(w, "Something went wrong", http.StatusInternalServerError)
		/*log.Printf("error decoding parameters: %s", err)

		// If there's an error, we still send out a response
		w.WriteHeader(http.StatusInternalServerError) // Code 500

		errText := generateErrorResp("Something went wrong")
		w.Write(errText)*/
		return
	}

	// Testing if the chirp is too long
	if len(chirp.Body) == 0 {
		respondError(w, "Chirp malformed", http.StatusBadRequest)
		/*w.WriteHeader(http.StatusBadRequest) // Code 400

		errBody := ChirpErrorResponse{
			Error: "Chirp malformed",
		}
		dat, err := json.Marshal(errBody)
		if err != nil {
			log.Printf("Error marshalling JSON: %s", err)
			return
		}
		w.Write(dat)*/
		return

	} else if len(chirp.Body) > 140 {
		respondError(w, "Chirp is too long", http.StatusBadRequest)
		/*w.WriteHeader(http.StatusBadRequest) // Code 400

		errBody := ChirpErrorResponse{
			Error: "Chirp is too long",
		}
		dat, err := json.Marshal(errBody)
		if err != nil {
			log.Printf("Error marshalling JSON: %s", err)
			return
		}
		w.Write(dat)*/
		return
	} else {

		chirp.Body = replaceProfane(chirp.Body)

		ccparams := database.CreateChirpParams{
			Body:   chirp.Body,
			UserID: chirp.UserID,
		}

		_, err := cfg.db.CreateChirp(r.Context(), ccparams)
		if err != nil {
			log.Printf("Error putting chirp into database: %v", err)
			respondError(w, "Couldn't process the chirp into database", http.StatusBadRequest)
			/*w.WriteHeader(http.StatusBadRequest) // Code 400
			errBody := ChirpErrorResponse{
				Error: "Couldn't process the chirp into database.",
			}
			dat, err := json.Marshal(errBody)
			if err != nil {
				log.Printf("Error marshalling JSON: %s", err)
				return
			}
			w.Write(dat)*/
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
	returnedChirps, err := cfg.db.GetChirps(r.Context())
	if err != nil {
		log.Printf("Error getting chirps from the database: %s", err)
		respondError(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	chirps := []chirp{}

	for _, item := range returnedChirps {
		chirps = append(chirps, chirp{
			ID:        item.ID,
			CreatedAt: item.CreatedAt,
			UpdatedAt: item.UpdatedAt,
			Body:      item.Body,
			UserID:    item.UserID,
		})
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
