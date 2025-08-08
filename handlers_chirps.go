package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
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
		Body   string    `json:"body"`
		UserID uuid.UUID `json:"user_id"`
	}
	params := parameters{}

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&params)
	if err != nil {
		decodeErr := fmt.Sprintf("Error decoding JSON: %v", err)
		helperResponseError(w, http.StatusBadRequest, decodeErr)
		return
	}
	if len(params.Body) > maxChirpLength {
		lengthErr := "Chirp is too long"
		helperResponseError(w, http.StatusBadRequest, lengthErr)
		return
	}
	chirp, err := cfg.db.CreateChirp(r.Context(), database.CreateChirpParams{
		Body:   helperCleanBody(params.Body),
		UserID: params.UserID,
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
		getChirpErr := fmt.Sprintf("Error retrieving all chirps: %v", err)
		helperResponseError(w, http.StatusInternalServerError, getChirpErr)
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
