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

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

func (cfg *apiConfig) handlerCreateChirp(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}
	params := parameters{}

	// check token
	invalidErr := "Invalid token"
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		helperResponseError(w, http.StatusUnauthorized, invalidErr)
		return
	}
	validID, err := auth.ValidateJWT(token, cfg.secret)
	if err != nil {
		helperResponseError(w, http.StatusUnauthorized, invalidErr)
		return
	}

	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&params)
	if err != nil {
		decodeErr := fmt.Sprintf("Error decoding JSON: %v", err)
		helperResponseError(w, http.StatusBadRequest, decodeErr)
		return
	}

	validBody, err := helperValidateBody(params.Body)
	if err != nil {
		validateBodyErr := fmt.Sprintf("Invalid chirp: %v", err)
		helperResponseError(w, http.StatusBadRequest, validateBodyErr)
	}
	chirp, err := cfg.db.CreateChirp(r.Context(), database.CreateChirpParams{
		Body:   validBody,
		UserID: validID,
	})
	if err != nil {
		createChirpErr := fmt.Sprintf("Error creating chirp: %v", err)
		helperResponseError(w, http.StatusInternalServerError, createChirpErr)
		return
	}
	helperResponseJSON(w, http.StatusCreated, Chirp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	})
}

func (cfg *apiConfig) handlerGetChirps(w http.ResponseWriter, r *http.Request) {
	dbChirps, err := cfg.db.GetChirps(r.Context())
	if err != nil {
		getChirpsErr := fmt.Sprintf("Error retrieving all chirps: %v", err)
		helperResponseError(w, http.StatusInternalServerError, getChirpsErr)
		return
	}
	chirps := []Chirp{}
	for _, dbChirp := range dbChirps {
		chirps = append(chirps, Chirp{
			ID:        dbChirp.ID,
			CreatedAt: dbChirp.CreatedAt,
			UpdatedAt: dbChirp.UpdatedAt,
			Body:      dbChirp.Body,
			UserID:    dbChirp.UserID,
		})
	}
	helperResponseJSON(w, http.StatusOK, chirps)
}

func (cfg *apiConfig) handlerGetAChirp(w http.ResponseWriter, r *http.Request) {
	chirpID, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		idErr := fmt.Sprintf("Invalid ID: %v", err)
		helperResponseError(w, http.StatusBadRequest, idErr)
		return
	}
	chirp, err := cfg.db.GetAChirp(r.Context(), chirpID)
	if err != nil {
		getAChirpErr := fmt.Sprintf("Error retrieving chirp: %v", err)
		helperResponseError(w, http.StatusNotFound, getAChirpErr)
		return
	}
	helperResponseJSON(w, http.StatusOK, Chirp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	})
}

func (cfg *apiConfig) handlerDeleteAChirp(w http.ResponseWriter, r *http.Request) {
	chirpID, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		idErr := fmt.Sprintf("Invalid ID: %v", err)
		helperResponseError(w, http.StatusBadRequest, idErr)
		return
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

	chirp, err := cfg.db.GetAChirp(r.Context(), chirpID)
	if err != nil {
		getAChirpErr := fmt.Sprintf("Error retrieving chirp: %v", err)
		helperResponseError(w, http.StatusNotFound, getAChirpErr)
		return
	}

	// verify chirp owner
	if chirp.UserID != user {
		userErr := "User not allowed to delete chirp"
		helperResponseError(w, http.StatusForbidden, userErr)
		return
	}

	// delete chirp
	err = cfg.db.DeleteAChirp(r.Context(), chirpID)
	if err != nil {
		deleteAChirpErr := fmt.Sprintf("Error deleting chirp: %v", err)
		helperResponseError(w, http.StatusInternalServerError, deleteAChirpErr)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
