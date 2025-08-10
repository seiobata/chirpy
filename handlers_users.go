package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/seiobata/chirpy/internal/auth"
	"github.com/seiobata/chirpy/internal/database"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

func (cfg *apiConfig) handlerCreateUser(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	params := parameters{}

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&params)
	if err != nil {
		decodeErr := fmt.Sprintf("Error decoding JSON: %v", err)
		helperResponseError(w, http.StatusBadRequest, decodeErr)
		return
	}
	hashedPass, err := auth.HashPassword(params.Password)
	if err != nil {
		hashErr := fmt.Sprintf("Error hashing password: %v", err)
		helperResponseError(w, http.StatusInternalServerError, hashErr)
		return
	}
	user, err := cfg.db.CreateUser(r.Context(), database.CreateUserParams{
		Email:          params.Email,
		HashedPassword: hashedPass,
	})
	if err != nil {
		createUserErr := fmt.Sprintf("Error creating user: %v", err)
		helperResponseError(w, http.StatusInternalServerError, createUserErr)
		return
	}
	helperResponseJSON(w, http.StatusCreated, User{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	})
}

func (cfg *apiConfig) handlerUserLogin(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	params := parameters{}

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&params)
	if err != nil {
		decodeErr := fmt.Sprintf("Error decoding JSON: %v", err)
		helperResponseError(w, http.StatusBadRequest, decodeErr)
		return
	}

	// check email and password
	invalidErr := "Incorrect email or password"
	user, err := cfg.db.GetUser(r.Context(), params.Email)
	if err != nil {
		helperResponseError(w, http.StatusUnauthorized, invalidErr)
		return
	}
	err = auth.CheckPasswordHash(params.Password, user.HashedPassword)
	if err != nil {
		helperResponseError(w, http.StatusUnauthorized, invalidErr)
		return
	}
	helperResponseJSON(w, http.StatusOK, User{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	})
}
