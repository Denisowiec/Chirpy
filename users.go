package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/Denisowiec/Chirpy/internal/auth"
	"github.com/Denisowiec/Chirpy/internal/database"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerCreateUser(w http.ResponseWriter, r *http.Request) {
	type createUserRequest struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	decoder := json.NewDecoder(r.Body)
	reqBody := createUserRequest{}

	w.Header().Set("Content-Type", "application/json")

	if err := decoder.Decode(&reqBody); err != nil {
		log.Printf("Error decoding parameters: %s", err)
		respondError(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	if reqBody.Password == "" {
		respondError(w, "No password provided", http.StatusBadRequest)
		return
	}

	hashedPassword, err := auth.HashPassword(reqBody.Password)
	if err != nil {
		log.Printf("Error hashing password: %s", err)
		respondError(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	crUsParams := database.CreateUserParams{
		Email:          reqBody.Email,
		HashedPassword: hashedPassword,
	}

	user, err := cfg.db.CreateUser(r.Context(), crUsParams)
	if err != nil {
		log.Printf("Error creating user: %s", err)
		respondError(w, "Could not create user", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated) // Code 201

	// We sent back the user's info, but withouot the password hash
	user.HashedPassword = ""

	dat, err := json.Marshal(user)
	if err != nil {
		log.Printf("Error marshalling json: %s", err)
		w.Write([]byte{})
		return
	}
	w.Write(dat)
}

func (cfg *apiConfig) handleLogin(w http.ResponseWriter, r *http.Request) {
	type loginRequest struct {
		Email            string `json:"email"`
		Password         string `json:"password"`
		ExpiresInSeconds int    `json:"expires_in_seconds"`
	}
	type loginResponse struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Email     string    `json:"email"`
		Token     string    `json:"token"`
	}

	decoder := json.NewDecoder(r.Body)
	reqBody := loginRequest{}

	if err := decoder.Decode(&reqBody); err != nil {
		log.Printf("Error decoding parameters: %s", err)
		respondError(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	if reqBody.Email == "" {
		respondError(w, "No e-mail provided", http.StatusBadRequest)
		return
	}
	if reqBody.Password == "" {
		respondError(w, "No password provided", http.StatusBadRequest)
		return
	}

	// We set up an expiration time for authorization
	var expiresIn time.Duration
	if reqBody.ExpiresInSeconds == 0 || reqBody.ExpiresInSeconds > 3600 {
		expiresIn = 3600 * time.Second
	} else {
		expiresIn = time.Duration(reqBody.ExpiresInSeconds) * time.Second
	}

	user, err := cfg.db.GetUserByEmail(r.Context(), reqBody.Email)
	if err != nil {
		log.Printf("Error looking up user %s in database: %s", reqBody.Email, err)
		respondError(w, "User not found", http.StatusUnauthorized)
		return
	}

	match, err := auth.CheckPasswordHash(reqBody.Password, user.HashedPassword)
	if err != nil {
		log.Printf("Error comparing password to hash: %s", err)
		respondError(w, "Something went wrong", http.StatusInternalServerError)
	}
	if !match {
		respondError(w, "Password incorrect", http.StatusUnauthorized)
		return
	}
	token, err := auth.MakeJWT(user.ID, cfg.jwtSecretCode, expiresIn)

	if err != nil {
		log.Printf("Error generating JWT: %s", err)
		respondError(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	resp := loginResponse{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
		Token:     token,
	}

	dat, err := json.Marshal(resp)
	if err != nil {
		log.Printf("Error marshalling json: %s", err)
		w.Write([]byte{})
	}
	w.Write(dat)
}
