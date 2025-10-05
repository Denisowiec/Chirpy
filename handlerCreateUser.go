package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/Denisowiec/Chirpy/internal/database"
	"github.com/google/uuid"
)

type createUserRequest struct {
	Email string `json:"email"`
}

type createUserResponse struct {
	Id        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

func (c *createUserResponse) makeFromDBReturn(u database.User) {
	c.Id = u.ID
	c.CreatedAt = u.CreatedAt
	c.UpdatedAt = u.UpdatedAt
	c.Email = u.Email
}

func (cfg *apiConfig) handlerCreateUser(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	reqBody := createUserRequest{}

	w.Header().Set("Content-Type", "application/json")

	if err := decoder.Decode(&reqBody); err != nil {
		log.Printf("Error decoding parameters: %s", err)

		w.WriteHeader(http.StatusInternalServerError) // Code 500

		errText := generateErrorResp("Something went wrong")
		w.Write(errText)
		return
	}

	dbUser, err := cfg.db.CreateUser(r.Context(), reqBody.Email)
	if err != nil {
		log.Printf("Error creating user: %s", err)

		w.WriteHeader(http.StatusBadRequest) // Code 400

		errText := generateErrorResp("Could not create user")
		w.Write(errText)
		return
	}
	user := createUserResponse{}
	user.makeFromDBReturn(dbUser)

	w.WriteHeader(http.StatusCreated) // Code 201

	dat, err := json.Marshal(user)
	if err != nil {
		log.Printf("Error marshalling json: %s", err)
	}
	w.Write(dat)
}
