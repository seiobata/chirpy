package main

import (
	"fmt"
	"net/http"

	"github.com/seiobata/chirpy/internal/auth"
)

func (cfg *apiConfig) handlerRefreshAccessToken(w http.ResponseWriter, r *http.Request) {
	type response struct {
		AccessToken string `json:"token"`
	}

	// validate refresh token
	invalidErr := "Token is invalid or expired"
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		helperResponseError(w, http.StatusUnauthorized, invalidErr)
		return
	}
	user, err := cfg.db.GetUserFromRefreshToken(r.Context(), refreshToken)
	if err != nil {
		helperResponseError(w, http.StatusUnauthorized, invalidErr)
		return
	}

	// generate new access token
	accessToken, err := auth.MakeJWT(user.ID, cfg.secret, accessTkExp)
	if err != nil {
		makeJWTErr := fmt.Sprintf("Error making JWT token: %v", err)
		helperResponseError(w, http.StatusInternalServerError, makeJWTErr)
		return
	}

	helperResponseJSON(w, http.StatusOK, response{
		AccessToken: accessToken,
	})
}

func (cfg *apiConfig) handlerRevokeRefreshToken(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		tokenErr := fmt.Sprintf("Header missing token: %v", err)
		helperResponseError(w, http.StatusBadRequest, tokenErr)
		return
	}

	_, err = cfg.db.RevokeRefreshToken(r.Context(), token)
	if err != nil {
		revokeErr := fmt.Sprintf("Error revoking refresh token: %v", err)
		helperResponseError(w, http.StatusInternalServerError, revokeErr)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
