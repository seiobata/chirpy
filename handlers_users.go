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
	ID          uuid.UUID `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Email       string    `json:"email"`
	IsChirpyRed bool      `json:"is_chirpy_red"`
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
		ID:          user.ID,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
		Email:       user.Email,
		IsChirpyRed: user.IsChirpyRed,
	})
}

func (cfg *apiConfig) handlerUpdateUser(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	// validate token
	invalidErr := "Token is invalid or expired"
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		helperResponseError(w, http.StatusUnauthorized, invalidErr)
		return
	}
	user, err := auth.ValidateJWT(token, cfg.secret)
	if err != nil {
		helperResponseError(w, http.StatusUnauthorized, invalidErr)
		return
	}

	// decode request parameters
	params := parameters{}
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&params)
	if err != nil {
		decodeErr := fmt.Sprintf("Error decoding JSON: %v", err)
		helperResponseError(w, http.StatusInternalServerError, decodeErr)
		return
	}

	// hash password
	password, err := auth.HashPassword(params.Password)
	if err != nil {
		hashErr := fmt.Sprintf("Error hashing password: %v", err)
		helperResponseError(w, http.StatusInternalServerError, hashErr)
		return
	}

	// update user email and password
	dbUser, err := cfg.db.UpdateUser(r.Context(), database.UpdateUserParams{
		ID:             user,
		Email:          params.Email,
		HashedPassword: password,
	})
	if err != nil {
		updateErr := fmt.Sprintf("Error updating user: %v", err)
		helperResponseError(w, http.StatusInternalServerError, updateErr)
		return
	}

	// request successful; returning user
	helperResponseJSON(w, http.StatusOK, User{
		ID:          dbUser.ID,
		CreatedAt:   dbUser.CreatedAt,
		UpdatedAt:   dbUser.UpdatedAt,
		Email:       dbUser.Email,
		IsChirpyRed: dbUser.IsChirpyRed,
	})
}

func (cfg *apiConfig) handlerUserLogin(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	type response struct {
		User
		AccessToken  string `json:"token"`
		RefreshToken string `json:"refresh_token"`
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

	// generate access token
	accessToken, err := auth.MakeJWT(user.ID, cfg.secret, accessTkExp)
	if err != nil {
		makeJWTErr := fmt.Sprintf("Error generating JWT token: %v", err)
		helperResponseError(w, http.StatusInternalServerError, makeJWTErr)
		return
	}

	// generate refresh token
	refreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		makeRefreshTkErr := fmt.Sprintf("Error generating refresh token: %v", err)
		helperResponseError(w, http.StatusInternalServerError, makeRefreshTkErr)
		return
	}

	// add refresh token to refresh_tokens database
	_, err = cfg.db.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
		UserID:    user.ID,
		Token:     refreshToken,
		ExpiresAt: time.Now().UTC().Add(refreshTkExp),
	})
	if err != nil {
		dbRefreshTkErr := fmt.Sprintf("Error adding refresh token to database: %v", err)
		helperResponseError(w, http.StatusInternalServerError, dbRefreshTkErr)
		return
	}

	helperResponseJSON(w, http.StatusOK, response{
		User: User{
			ID:          user.ID,
			CreatedAt:   user.CreatedAt,
			UpdatedAt:   user.UpdatedAt,
			Email:       user.Email,
			IsChirpyRed: user.IsChirpyRed,
		},
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})
}

func (cfg *apiConfig) handlerUpgradeUserToRed(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Event string `json:"event"`
		Data  struct {
			UserID string `json:"user_id"`
		} `json:"data"`
	}

	params := parameters{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&params)
	if err != nil {
		decodeErr := fmt.Sprintf("Error decoding JSON: %v", err)
		helperResponseError(w, http.StatusBadRequest, decodeErr)
		return
	}

	// get user ID
	user, err := uuid.Parse(params.Data.UserID)
	if err != nil {
		idErr := fmt.Sprintf("Error parsing ID: %v", err)
		helperResponseError(w, http.StatusBadRequest, idErr)
		return
	}

	// filter for user.upgraded event
	if params.Event != "user.upgraded" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// upgrade user in database
	err = cfg.db.UserIsChirpyRed(r.Context(), user)
	if err != nil {
		isChirpyRedErr := fmt.Sprintf("Error updating user: %v", err)
		helperResponseError(w, http.StatusNotFound, isChirpyRedErr)
		return
	}

	// successful request; return 'no content' header
	w.WriteHeader(http.StatusNoContent)
}
