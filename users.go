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

func (cfg *apiConfig) handleUpdateUser(w http.ResponseWriter, r *http.Request) {
	// Handle PUT request on users, allows updating email and password
	type UpdateUserRequest struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	type UpdateUserResponse struct {
		ID        uuid.UUID `json:"id"`
		Email     string    `json:"email"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
	}

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

	decoder := json.NewDecoder(r.Body)
	reqBody := UpdateUserRequest{}
	if err := decoder.Decode(&reqBody); err != nil {
		log.Printf("Error decoding parameters: %s", err)
		respondError(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	newPassword, err := auth.HashPassword(reqBody.Password)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		respondError(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	UUParams := database.UpdateUserParams{
		ID:             inUID,
		Email:          reqBody.Email,
		HashedPassword: newPassword,
	}

	user, err := cfg.db.UpdateUser(r.Context(), UUParams)

	if err != nil {
		log.Printf("Error updating user in database: %s", err)
		respondError(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	respBody := UpdateUserResponse{
		ID:        user.ID,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	dat, err := json.Marshal(respBody)
	if err != nil {
		log.Printf("Error marshalling json: %s", err)
		dat = []byte{}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(dat)
}

func (cfg *apiConfig) handleLogin(w http.ResponseWriter, r *http.Request) {
	type loginRequest struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	type loginResponse struct {
		ID            uuid.UUID `json:"id"`
		CreatedAt     time.Time `json:"created_at"`
		UpdatedAt     time.Time `json:"updated_at"`
		Email         string    `json:"email"`
		Token         string    `json:"token"`
		Refresh_token string    `json:"refresh_token"`
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
	// We set up an expiration time for the JWT. It's one hour by default
	expiresIn := 3600 * time.Second
	token, err := auth.MakeJWT(user.ID, cfg.jwtSecretCode, expiresIn)

	if err != nil {
		log.Printf("Error generating JWT: %s", err)
		respondError(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	// We create a refresh token, lasting 60 days
	refToken := auth.MakeRefreshToken()
	setRefParams := database.SetRefTokenParams{
		Token:     refToken,
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(time.Hour * 24 * 60),
	}
	_, err = cfg.db.SetRefToken(r.Context(), setRefParams)
	if err != nil {
		log.Printf("Error recording the refresh token in the database: %s", err)
		respondError(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	resp := loginResponse{
		ID:            user.ID,
		CreatedAt:     user.CreatedAt,
		UpdatedAt:     user.UpdatedAt,
		Email:         user.Email,
		Token:         token,
		Refresh_token: refToken,
	}

	dat, err := json.Marshal(resp)
	if err != nil {
		log.Printf("Error marshalling json: %s", err)
		w.Write([]byte{})
	}
	w.Write(dat)
}

func (cfg *apiConfig) handleRefresh(w http.ResponseWriter, r *http.Request) {
	// This function refreshes the login credentials

	inRefToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("Error extracting token from header: %s", err)
		respondError(w, "Authentification failed", http.StatusUnauthorized)
		return
	}

	token, err := cfg.db.GetRefToken(r.Context(), inRefToken)
	if err != nil {
		log.Printf("Error getting token information from database")
		respondError(w, "Authentification failed", http.StatusUnauthorized)
		return
	}
	// If revoked_at is not null, that means the token has been revoked
	if token.RevokedAt.Valid {
		log.Printf("Token revoked")
		respondError(w, "Authentification failed", http.StatusUnauthorized)
		return
	}

	// If refreshtoken given is valid we offer an access token
	jwt, err := auth.MakeJWT(token.UserID, cfg.jwtSecretCode, time.Hour)
	if err != nil {
		log.Printf("Error creating access token for user: %s", err)
		respondError(w, "Something went wrong.", http.StatusInternalServerError)
		return
	}

	var respBody struct {
		Token string `json:"token"`
	}
	respBody.Token = jwt

	dat, err := json.Marshal(respBody)
	if err != nil {
		log.Printf("Error marshalling json: %s", err)
		dat = []byte{}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(dat)

}

func (cfg *apiConfig) handleRevoke(w http.ResponseWriter, r *http.Request) {
	// This function revokes a refresh token
	inRefToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("Error extracting token from header: %s", err)
		respondError(w, "Authentification failed", http.StatusUnauthorized)
		return
	}

	token, err := cfg.db.GetRefToken(r.Context(), inRefToken)
	if err != nil {
		log.Printf("Error getting token information from database")
		respondError(w, "Authentification failed", http.StatusUnauthorized)
		return
	}
	// If revoked_at is not null, that means the token has been revoked
	if token.RevokedAt.Valid {
		log.Printf("Token already revoked")
		respondError(w, "Authentification failed", http.StatusUnauthorized)
		return
	}

	// If refreshtoken given is valid revoke it
	_, err = cfg.db.RevokeToken(r.Context(), token.Token)
	if err != nil {
		log.Printf("Error revoking token: %s", err)
		respondError(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
